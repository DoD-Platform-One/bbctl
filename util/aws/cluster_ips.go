package aws

// ClusterIP type
type ClusterIP struct {
	IP            *string
	ReservationID *string
	InstanceID    *string
	IsPublic      bool
}

// SortedClusterIPs type
type SortedClusterIPs struct {
	PublicIPs  []ClusterIP
	PrivateIPs []ClusterIP
}
