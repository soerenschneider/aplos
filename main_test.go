package main

import "testing"

func TestConf_Validate(t *testing.T) {
	type fields struct {
		Address              string
		Directory            string
		TlsCertFile          string
		TlsKeyFile           string
		HealthHandlerPattern string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "happy case",
			fields: fields{
				Address:              ":8080",
				Directory:            "/tmp",
				HealthHandlerPattern: "/_handler",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			fields: fields{
				Address:              "8080",
				Directory:            "/tmp",
				HealthHandlerPattern: "/_handler",
			},
			wantErr: true,
		},
		{
			name: "non existent directory",
			fields: fields{
				Address:              ":8080",
				Directory:            "/nonexistentfoldertmp",
				HealthHandlerPattern: "/_handler",
			},
			wantErr: true,
		},
		{
			name: "invalid handler pattern",
			fields: fields{
				Address:              ":8080",
				Directory:            "/tmp",
				HealthHandlerPattern: "_handler",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conf{
				Address:             tt.fields.Address,
				Directory:           tt.fields.Directory,
				TlsCertFile:         tt.fields.TlsCertFile,
				TlsKeyFile:          tt.fields.TlsKeyFile,
				HealthcheckEndpoint: tt.fields.HealthHandlerPattern,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
