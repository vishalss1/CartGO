package auth

const (
	// HeaderUserID is the internal header used to propagate the authenticated User ID
	HeaderUserID = "X-User-ID"
	
	// HeaderUserRole is the internal header used to propagate the authenticated User Role
	HeaderUserRole = "X-User-Role"
)

const (
	RoleAdmin           = "ADMIN"
	RoleCustomer        = "CUSTOMER"
	RoleWarehouseStaff  = "WAREHOUSE_STAFF"
	RoleDeliveryPartner = "DELIVERY_PARTNER"
	RoleSupportAgent    = "SUPPORT_AGENT"
	RoleServiceOrder    = "SERVICE_ORDER"
)
