; zone file fragment for {{.ZoneName}}

;$TTL 600

$ORIGIN {{.ZoneName}}
; SOA record
; owner-name ttl class rr      name-server      email-addr  (sn ref ret ex min)
@                 IN   SOA     ns1.{{.ZoneName}}.   root.{{.ZoneName}}. (
			2017031088 ; sn = serial number
			3600       ; ref = refresh = 20m
			180        ; uret = update retry = 1m
			1209600    ; ex = expiry = 2w
			10800      ; nx = nxdomain ttl = 3h
			)
; type syntax
; host ttl class type data
{{range $k,$rr := .RRList}}
{{$rr.Data}}
{{end}}
