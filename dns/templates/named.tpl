options {
	directory "{{.ConfigPath}}";
	pid-file "named.pid";
	allow-new-zones yes;
	allow-query {any;};
	dnssec-enable no;
	dnssec-validation no;
	{{if .Forward}}forward {{.Forward.ForwardType}};
	forwarders{ {{range $k, $ip := .Forward.IPs}}{{$ip}};{{end}} };{{end}}{{range $k, $dns64:= .DNS64s}}
        dns64 {{$dns64.Prefix}} {
        clients { {{$dns64.ClientACLName}}; };
        mapped { {{$dns64.AAddressACLName}}; };
        exclude { {{$dns64.Prefix}}; };
        suffix ::;
        };{{end}}{{if .IPBlackHole}}
	BlackHole{ {{range $k,$v := .IPBlackHole.ACLNames}}{{$v}}; {{end}}};{{end}}{{if .Concu}}
	recursive-clients {{.Concu.RecursiveClients}};
	fetches-per-zone {{.Concu.FetchesPerZone}};{{end}}
};
key key1 {
    algorithm hmac-md5;
    secret "bGlua2luZ19lbmNy";
};

{{range $k, $view := .Views}}
view "{{$view.Name}}" {
	match-clients {
	{{range $kk, $acl := $view.ACLs}}{{$acl.Name}};{{end}}
	key key1;
	};
	allow-update {key key1;};{{range $i, $zone := $view.Zones}}{{if $zone.Forwarder}}
	zone "{{$zone.Name}}" { type forward; forward {{$zone.ForwardType}}; forwarders { {{range $ii,$ip := $zone.Forwarder.IPs}}{{$ip}}; {{end}}}; };{{end}}{{end}}{{range $k, $dns64:= .DNS64s}}
        dns64 {{$dns64.Prefix}} {
        clients { {{$dns64.ClientACLName}}; };
        mapped { {{$dns64.AAddressACLName}}; };
        exclude { {{$dns64.Prefix}}; };
        suffix ::;
        };{{end}}{{if $view.Redirect}}
	zone "." {
        type redirect;
        file "redirection/redirect_{{$view.Name}}";
        };{{end}}{{if $view.RPZ}}
	response-policy { zone "rpz" policy given; } max-policy-ttl 86400 qname-wait-recurse no ;
        zone "rpz" {type master; file "redirection/rpz_{{$view.Name}}"; allow-query {any;}; };{{end}}
};{{end}}

key "rndc-key" {
	algorithm hmac-sha256;
	secret "4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=";
};
controls {
        inet 127.0.0.1 port 953
        allow { 127.0.0.1; } keys { "rndc-key"; };
};
{{range $k, $acl := .ACLNames}}
include "/root/bindtest/{{$acl}}.conf";{{end}}

