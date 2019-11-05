package dns

type View struct {
	ID       string
	ViewName string
	ACLLIst  map[string]ACL
	ZoneList map[string]Zone
}
