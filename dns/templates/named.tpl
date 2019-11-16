options {
	directory "{{.ConfigPath}}";
	pid-file "named.pid";
	allow-new-zones yes;
	allow-query {any;};
};
key key1 {
    algorithm hmac-md5;
    secret "bGlua2luZ19lbmNy";
};

{{range $k, $view := .Views}}
view "{{$view.Name}}" {
	match-clients {	{{range $kk, $acl := $view.ACLs}}
	{{$acl.Name}};	{{end}}
	key key1;
	};
	allow-update { key key1;};
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
{{range $k, $view := .Views}}{{range $kk, $acl := $view.ACLs}}
include "/root/bindtest/{{$acl.Name}}.conf";{{end}}{{end}}

