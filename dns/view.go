package dns

type View struct {
	ID       string
	ViewName string
	ACLList  map[string]ACL
	ZoneList map[string]Zone
}
