package nskeyedarchiver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

// TestDecoderJson tests if real DTX nsKeyedArchiver plist can be decoded without error
// func TestDecoderJson(t *testing.T) {
// 	dat, err := ioutil.ReadFile("fixtures/payload_dump.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var payloads []string
// 	json.Unmarshal([]byte(dat), &payloads)
// 	for _, plistHex := range payloads {
// 		plistBytes, _ := hex.DecodeString(plistHex)
// 		_, err := Unarchive(plistBytes)
// 		assert.NoError(t, err)
// 	}
// }

func TestUnarchiveXml(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test one value",
			args: args{
				filename: "onevalue",
			},
			want:    "[true]",
			wantErr: false,
		},
		{
			name: "test all primitives",
			args: args{
				filename: "primitives",
			},
			want:    "[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42]",
			wantErr: false,
		},
		{
			name: "test arrays and sets",
			args: args{
				filename: "arrays",
			},
			want:    "[[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42],[true,\"Hello, World!\",42],[true],[42,true,\"Hello, World!\"]]",
			wantErr: false,
		},
		{
			name: "test nested arrays",
			args: args{
				filename: "nestedarrays",
			},
			want:    "[[[true],[42,true,\"Hello, World!\"]]]",
			wantErr: false,
		},
		{
			name: "test dictionaries",
			args: args{
				filename: "dict",
			},
			want:    "[{\"array\":[true,\"Hello, World!\",42],\"int\":1,\"string\":\"string\"}]",
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := ioutil.ReadFile("fixtures/" + tt.args.filename + ".xml")
			got, err := Unarchive(dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unarchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(convertToJSON(got), tt.want) {
				t.Errorf("Unarchive() = %v, want %v", convertToJSON(got), tt.want)
			}
		})
	}
}

func TestUnarchiveBin(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test one value",
			args: args{
				filename: "onevalue",
			},
			want:    "[true]",
			wantErr: false,
		},
		{
			name: "test all primitives",
			args: args{
				filename: "primitives",
			},
			want:    "[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42]",
			wantErr: false,
		},
		{
			name: "test arrays and sets",
			args: args{
				filename: "arrays",
			},
			want:    "[[1,1,1,1.5,\"YXNkZmFzZGZhZHNmYWRzZg==\",true,\"Hello, World!\",\"Hello, World!\",\"Hello, World!\",false,false,42],[true,\"Hello, World!\",42],[true],[42,true,\"Hello, World!\"]]",
			wantErr: false,
		},
		{
			name: "test nested arrays",
			args: args{
				filename: "nestedarrays",
			},
			want:    "[[[true],[42,true,\"Hello, World!\"]]]",
			wantErr: false,
		},
		{
			name: "test dictionaries",
			args: args{
				filename: "dict",
			},
			want:    "[{\"array\":[true,\"Hello, World!\",42],\"int\":1,\"string\":\"string\"}]",
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := ioutil.ReadFile("fixtures/" + tt.args.filename + ".bin")
			got, err := Unarchive(dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unarchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(convertToJSON(got), tt.want) {
				t.Errorf("Unarchive() = %v, want %v", convertToJSON(got), tt.want)
			}
		})
	}
}

func TestUnarchiveValidation(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "$archiver key is missing",
			args: args{
				filename: "missing_archiver",
			},
			wantErr: true,
		},
		{
			name: "$archiver is not nskeyedarchiver",
			args: args{
				filename: "wrong_archiver",
			},
			wantErr: true,
		},
		{
			name: "$top key is missing",
			args: args{
				filename: "missing_top",
			},
			wantErr: true,
		},
		{
			name: "$objects key is missing",
			args: args{
				filename: "missing_objects",
			},
			wantErr: true,
		},
		{
			name: "$version key is missing",
			args: args{
				filename: "missing_version",
			},
			wantErr: true,
		},
		{
			name: "$version is wrong",
			args: args{
				filename: "wrong_version",
			},
			wantErr: true,
		},
		{
			name: "plist is invalid",
			args: args{
				filename: "broken_plist",
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dat, err := ioutil.ReadFile("fixtures/" + tt.args.filename + ".bin")
			_, err = Unarchive(dat)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unarchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func convertToJSON(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(b)
}
