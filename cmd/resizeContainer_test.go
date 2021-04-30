package cmd

import (
	"testing"
)

func Test_checkValidContainerID(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, err := validateContainerID(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateContainerID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValid != tt.wantValid {
				t.Errorf("validateContainerID() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
		})
	}
}

func Test_getContainerSize(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name     string
		args     args
		wantSize string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSize, err := getContainerSize(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSize != tt.wantSize {
				t.Errorf("getContainerSize() gotSize = %v, want %v", gotSize, tt.wantSize)
			}
		})
	}
}

func Test_getRootContainerID(t *testing.T) {
	tests := []struct {
		name    string
		wantId  string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := getRootContainerID()
			if (err != nil) != tt.wantErr {
				t.Errorf("getRootContainerID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("getRootContainerID() gotId = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func Test_growContainer(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name        string
		args        args
		wantMessage string
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMessage, err := growContainer(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("growContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("growContainer() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}

func Test_growRootContainer(t *testing.T) {
	tests := []struct {
		name        string
		wantMessage string
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMessage, err := growRootContainer()
			if (err != nil) != tt.wantErr {
				t.Errorf("growRootContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMessage != tt.wantMessage {
				t.Errorf("growRootContainer() gotMessage = %v, want %v", gotMessage, tt.wantMessage)
			}
		})
	}
}

func Test_resizeContainer(t *testing.T) {
	type args struct {
		id   string
		size string
	}
	tests := []struct {
		name        string
		args        args
		wantNewSize string
		wantErr     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewSize, err := resizeContainer(tt.args.id, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("resizeContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNewSize != tt.wantNewSize {
				t.Errorf("resizeContainer() gotNewSize = %v, want %v", gotNewSize, tt.wantNewSize)
			}
		})
	}
}
