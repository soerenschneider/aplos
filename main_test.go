package main

import "testing"

func TestConf_Validate(t *testing.T) {
	type fields struct {
		Address              string
		Directory            string
		TlsCertFile          string
		TlsKeyFile           string
		HealthHandlerPattern string
		IdleTimeoutSec       int
		ReadTimeoutSec       int
		WriteTimeoutSec      int
		ReadHeaderTimeoutSec int
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
				IdleTimeoutSec:       30,
				ReadTimeoutSec:       30,
				WriteTimeoutSec:      30,
				ReadHeaderTimeoutSec: 30,
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			fields: fields{
				Address:              "8080",
				Directory:            "/tmp",
				HealthHandlerPattern: "/_handler",
				IdleTimeoutSec:       30,
				ReadTimeoutSec:       30,
				WriteTimeoutSec:      30,
				ReadHeaderTimeoutSec: 30,
			},
			wantErr: true,
		},
		{
			name: "non existent directory",
			fields: fields{
				Address:              ":8080",
				Directory:            "/nonexistentfoldertmp",
				HealthHandlerPattern: "/_handler",
				IdleTimeoutSec:       30,
				ReadTimeoutSec:       30,
				WriteTimeoutSec:      30,
				ReadHeaderTimeoutSec: 30,
			},
			wantErr: true,
		},
		{
			name: "invalid handler pattern",
			fields: fields{
				Address:              ":8080",
				Directory:            "/tmp",
				HealthHandlerPattern: "_handler",
				IdleTimeoutSec:       30,
				ReadTimeoutSec:       30,
				WriteTimeoutSec:      30,
				ReadHeaderTimeoutSec: 30,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Conf{
				Address:              tt.fields.Address,
				Directory:            tt.fields.Directory,
				TlsCertFile:          tt.fields.TlsCertFile,
				TlsKeyFile:           tt.fields.TlsKeyFile,
				HealthcheckEndpoint:  tt.fields.HealthHandlerPattern,
				IdleTimeoutSec:       tt.fields.IdleTimeoutSec,
				ReadTimeoutSec:       tt.fields.ReadTimeoutSec,
				WriteTimeoutSec:      tt.fields.WriteTimeoutSec,
				ReadHeaderTimeoutSec: tt.fields.ReadHeaderTimeoutSec,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
