options {
	directory "{{.ConfigPath}}/";
	pid-file "named.pid";
	allow-new-zones yes;
	allow-query {any;};
};
{{range $k, $view := .ViewList}}
view "{{$view.ViewName}}" {
	match-clients {	{{range $kk, $acl := $view.ACLList}}
	{{$acl.Name}};	{{end}}
	};{{range $kkk, $zone := $view.ZoneList}}	
	zone "{{$zone.ZoneName}}" {
	type master;
	file "{{$zone.ZoneFileName}}.zone";
	};	{{end}}
};{{end}}

view "default" {
        match-clients {
        any;
        };
};
key "rndc-key" {
	algorithm hmac-sha256;
	secret "4WqnJgCtpG8dPHDCBjwyQKtOzAPgiS+Iah5KN4xeq/U=";
};
controls {
        inet 127.0.0.1 port 953
        allow { 127.0.0.1; } keys { "rndc-key"; };
};
{{range $k, $view := .ViewList}}{{range $kk, $acl := $view.ACLList}}
include "/root/bindtest/{{$acl.Name}}.conf";{{end}}{{end}}

