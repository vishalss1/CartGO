import { useState, useEffect, useCallback, useRef } from "react";
import { authenticatedRequest } from "../utils/api";
import { SYSTEM_STATES, SYSTEM_INVARIANTS } from "../utils/systemModel";
import { logger } from "../utils/logger";

export function useSystemOrder(orderId, token) {
  const [data, setData] = useState({ order: null, payment: null, delivery: null });
  const [derivedState, setDerivedState] = useState(SYSTEM_STATES.CREATED);
  const [logs, setLogs] = useState([]);
  const [inconsistency, setInconsistency] = useState(null);
  const [loading, setLoading] = useState(true);
  
  const prevStateRef = useRef(null);

  const fetchSystemData = useCallback(async () => {
    if (!orderId || !token) return;

    try {
      const [order, paymentResult, deliveryResult] = await Promise.allSettled([
        authenticatedRequest(`/orders/${orderId}`, token),
        authenticatedRequest(`/payments/order/${orderId}`, token),
        authenticatedRequest(`/deliveries/order/${orderId}`, token),
      ]);

      const newState = {
        order: order.status === "fulfilled" ? order.value : null,
        payment: paymentResult.status === "fulfilled" ? paymentResult.value : null,
        delivery: deliveryResult.status === "fulfilled" ? deliveryResult.value : null,
      };

      setData(newState);
      deriveAndTransition(newState);
    } catch (err) {
      logger.error("System Fetch Error:", err);
    } finally {
      setLoading(false);
    }
  }, [orderId, token]);

  const deriveAndTransition = (state) => {
    let nextState = SYSTEM_STATES.CREATED;

    // 1. Terminal Check (DELIVERED is immutable)
    if (prevStateRef.current === SYSTEM_STATES.DELIVERED) {
      nextState = SYSTEM_STATES.DELIVERED;
    } 
    // 2. Logic Chain
    else if (state.delivery?.status === "DELIVERED") {
      nextState = SYSTEM_STATES.DELIVERED;
    } else if (state.delivery?.status === "PICKED_UP") {
      // Synthetic Transit logic
      nextState = SYSTEM_STATES.IN_TRANSIT;
    } else if (state.delivery?.status === "PENDING" && state.delivery?.delivery_person_id) {
      nextState = SYSTEM_STATES.DELIVERY_ASSIGNED;
    } else if (state.order?.status === "CONFIRMED") {
      if (!state.delivery) {
        nextState = SYSTEM_STATES.RESOLVING_DELIVERY_ASSIGNMENT;
      } else {
        nextState = SYSTEM_STATES.INVENTORY_RESERVED;
      }
    } else if (state.payment?.status === "SUCCESS") {
      if (state.order?.status === "PENDING" || state.order?.status === "FAILED") {
        nextState = SYSTEM_STATES.RESOLVING_ORDER_CONFIRMATION;
        // Automatic Resolution Trigger
        // 2. Synchronize Order Fulfillment
        authenticatedRequest(`/orders/${orderId}/confirm-after-payment`, token, { method: "POST" })
             .then(() => fetchSystemData())
             .catch(err => logger.error("Auto-Resolution failed:", err));
      } else {
        nextState = SYSTEM_STATES.CONFIRMED;
      }
    } else if (state.payment?.status === "FAILURE") {
      nextState = SYSTEM_STATES.PAYMENT_FAILED;
    } else if (state.order?.status === "PENDING") {
      nextState = SYSTEM_STATES.CREATED;
    }

    // 3. Invariant Checks
    const violation = SYSTEM_INVARIANTS.find(inv => !inv.rule(state, nextState));
    if (violation) {
      setInconsistency(violation);
      setDerivedState(SYSTEM_STATES.INCONSISTENT);
    } else {
      setInconsistency(null);
      
      // 4. Traceable Logging
      if (nextState !== prevStateRef.current) {
        const logEntry = {
          timestamp: new Date().toISOString(),
          from: prevStateRef.current,
          to: nextState,
          orderId: orderId,
          service: getServiceForState(nextState)
        };
        setLogs(prev => [...prev, logEntry]);
        prevStateRef.current = nextState;
        setDerivedState(nextState);
      }
    }
  };

  const getServiceForState = (state) => {
    if (state.includes("PAYMENT")) return "Payment Service";
    if (state.includes("DELIVERY") || state === "PICKED_UP" || state === "IN_TRANSIT") return "Delivery Service";
    if (state === "INVENTORY_RESERVED") return "Inventory Service";
    return "Order Service";
  };

  useEffect(() => {
    fetchSystemData();
    const interval = setInterval(fetchSystemData, 15000);
    return () => clearInterval(interval);
  }, [fetchSystemData]);

  return { 
    ...data, 
    derivedState, 
    logs, 
    inconsistency, 
    loading,
    refresh: fetchSystemData 
  };
}
