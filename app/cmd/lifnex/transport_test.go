package main

import (
	"testing"
)

func TestLoadTransportOptions(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		port      string
		want      transportOptions
		wantErr   string
	}{
		{name: "defaults", want: transportOptions{transport: stdioTransport, port: defaultPort}},
		{name: "http", transport: "HTTP", port: "9090", want: transportOptions{transport: httpTransport, port: 9090}},
		{name: "invalid transport", transport: "sse", wantErr: `failed to parse TRANSPORT "sse": unsupported value`},
		{name: "invalid port text", transport: "http", port: "abc", wantErr: `failed to parse PORT "abc": strconv.Atoi: parsing "abc": invalid syntax`},
		{name: "invalid port range", transport: "http", port: "65536", wantErr: `failed to validate PORT "65536": must be between 1 and 65535`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("TRANSPORT", test.transport)
			t.Setenv("PORT", test.port)

			got, err := loadTransportOptions()
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("loadTransportOptions() error = nil, want %q", test.wantErr)
				}
				if err.Error() != test.wantErr {
					t.Errorf("loadTransportOptions() error = %q, want %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadTransportOptions() error = %v, want nil", err)
			}
			if got != test.want {
				t.Errorf("loadTransportOptions() = %+v, want %+v", got, test.want)
			}
		})
	}
}
