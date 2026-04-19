#!/bin/bash
# Inventory Sync Utility
# This script ensures every product in the product-service has a corresponding inventory record.

echo "Fetching product IDs..."
PRODUCT_IDS=$(docker exec -i cartgo-db psql -U cartgo_user -d cartgo_product_db -t -c "SELECT id FROM products;")

echo "Ensuring inventory records..."
for ID in $PRODUCT_IDS; do
    echo "Processing product: $ID"
    # Call the internal AdjustStock logic via curl to the gateway (as admin)
    # Status 200/204 means success (either updated or created via upsert)
    docker exec -i api-gateway wget -qO- --header="X-User-Role: SERVICE_SYNC" --header="X-User-ID: 00000000-0000-0000-0000-000000000000" --post-data='{"adjustment": 0}' http://inventory-service:8083/api/v1/inventory/$ID/adjust
done

echo "Sync complete."
