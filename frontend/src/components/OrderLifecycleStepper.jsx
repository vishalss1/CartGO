import { SYSTEM_STATES, SERVICE_OWNERSHIP } from "../utils/systemModel";

const STEPS = [
  { id: SYSTEM_STATES.CREATED, label: "Created" },
  { id: SYSTEM_STATES.PAYMENT_SUCCESS, label: "Payment" },
  { id: SYSTEM_STATES.CONFIRMED, label: "Confirmed" },
  { id: SYSTEM_STATES.INVENTORY_RESERVED, label: "Processed" },
  { id: SYSTEM_STATES.DELIVERY_ASSIGNED, label: "Assigned" },
  { id: SYSTEM_STATES.IN_TRANSIT, label: "Transit" },
  { id: SYSTEM_STATES.DELIVERED, label: "Delivered" },
];

export default function OrderLifecycleStepper({ currentState, inconsistency }) {
  const getStepStatus = (stepId, index) => {
    if (inconsistency) return "error";
    
    const currentIndex = STEPS.findIndex(s => s.id === currentState) || 0;
    const stepIndex = STEPS.findIndex(s => s.id === stepId);
    
    if (currentState === SYSTEM_STATES.PAYMENT_FAILED && stepId === SYSTEM_STATES.PAYMENT_SUCCESS) return "failed";
    if (stepId === currentState) return "active";
    if (stepIndex < currentIndex && currentIndex !== -1) return "completed";
    return "pending";
  };

  return (
    <div className="w-full space-y-8 py-10">
      <div className="relative flex justify-between">
        {/* Background Line */}
        <div className="absolute top-5 left-0 h-[2px] w-full bg-line" />
        
        {STEPS.map((step, idx) => {
          const status = getStepStatus(step.id, idx);
          
          return (
            <div key={step.id} className="relative z-10 flex flex-col items-center">
              <div 
                className={`flex h-10 w-10 items-center justify-center border-2 transition-all duration-500 scale-100
                  ${status === 'completed' ? 'bg-paper border-paper text-surface' : ''}
                  ${status === 'active' ? 'bg-surface border-paper text-paper animate-pulse scale-110 shadow-[0_0_15px_rgba(255,255,255,0.2)]' : ''}
                  ${status === 'failed' ? 'bg-red-500 border-red-500 text-white' : ''}
                  ${status === 'error' ? 'bg-red-600 border-red-600 text-white animate-bounce' : ''}
                  ${status === 'pending' ? 'bg-surface border-line text-muted' : ''}
                `}
              >
                {status === 'completed' ? (
                  <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                  </svg>
                ) : (
                  <span className="text-xs font-bold">{idx + 1}</span>
                )}
              </div>
              
              <div className="mt-4 text-center">
                <p className={`text-[10px] font-black uppercase tracking-widest ${status === 'active' ? 'text-paper' : 'text-muted'}`}>
                  {step.label}
                </p>
                {status === 'active' && (
                   <p className="mt-1 text-[8px] font-bold uppercase tracking-[0.2em] text-accent opacity-80">
                      {SERVICE_OWNERSHIP[currentState] || "Resolving..."}
                   </p>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Resolution Indicator */}
      {(currentState.includes("RESOLVING") || currentState === SYSTEM_STATES.PAYMENT_PENDING) && (
        <div className="flex items-center justify-center gap-3 rounded-sm border border-accent/20 bg-accent/5 py-3 text-[10px] font-bold uppercase tracking-widest text-accent">
          <svg className="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          System Consensus in Progress: {currentState.replaceAll("_", " ")}
        </div>
      )}

      {inconsistency && (
        <div className="flex flex-col items-center gap-2 rounded-sm border border-red-500/30 bg-red-500/10 p-4 text-center">
          <span className="text-xs font-black uppercase tracking-widest text-red-500">System Invariant Violation Detected</span>
          <p className="max-w-md text-[10px] leading-relaxed text-red-400">
             Warning: {inconsistency.name}. Theoretical state drift detected between distributed services.
             Error Reference: {inconsistency.errorCode}
          </p>
        </div>
      )}
    </div>
  );
}
