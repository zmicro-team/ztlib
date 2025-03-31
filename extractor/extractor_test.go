package extractor

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testHeaderExtractorType string

// nolint
func TestExtractors(t *testing.T) {
	ctx := context.WithValue(context.Background(), "Authorization", "Bearer abc123")
	ctx = context.WithValue(ctx, "Token", "abc123")
	request := &http.Request{
		Header: http.Header{
			"Authorization": []string{"Bearer xyz456"},
			"Token":         []string{"abc123"},
		},
	}

	headerExtractor := HeaderExtractor{"Token", "Authorization"}
	bearerExtractor := BearerExtractor{}
	authorizationExtractor := AuthorizationExtractor{}
	tokenExtractor := TokenExtractor{}

	// Test HeaderExtractor
	token, err := headerExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = headerExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	// Test BearerExtractor
	token, err = bearerExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = bearerExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "xyz456", token)

	// Test AuthorizationExtractor
	token, err = authorizationExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "Bearer abc123", token)

	token, err = authorizationExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "Bearer xyz456", token)

	// Test TokenExtractor
	token, err = tokenExtractor.Extract(ctx)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)

	token, err = tokenExtractor.ExtractRequest(request)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)
}

// nolint
func TestHeaderExtractor_Extract(t *testing.T) {
	var authorizationStr testHeaderExtractorType = "Authorization"
	ctx := context.WithValue(context.Background(), string(authorizationStr), "abc123")
	var tokenStr testHeaderExtractorType = "token"
	ctx = context.WithValue(ctx, tokenStr, "321abc")
	extractor := HeaderExtractor{string(authorizationStr), string(tokenStr)}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}

// nolint
func TestBearerExtractor_Extract(t *testing.T) {
	var authorizationStr testHeaderExtractorType = "Authorization"
	ctx := context.WithValue(context.Background(), string(authorizationStr), "Bearer abc123")
	extractor := BearerExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}

// nolint
func TestAuthorizationExtractor_Extract(t *testing.T) {
	var authorizationStr testHeaderExtractorType = "Authorization"
	ctx := context.WithValue(context.Background(), string(authorizationStr), "Bearer abc123")
	extractor := AuthorizationExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "Bearer abc123", token)
}

// nolint
func TestTokenExtractor_Extract(t *testing.T) {
	var tokenStr testHeaderExtractorType = "Token"
	ctx := context.WithValue(context.Background(), string(tokenStr), "abc123")
	extractor := TokenExtractor{}

	token, err := extractor.Extract(ctx)

	assert.Equal(t, nil, err)
	assert.Equal(t, "abc123", token)
}

// Test HeaderExtractor ExtractHeader
func TestHeaderExtractor_ExtractHeader(t *testing.T) {
	headerExtractor := HeaderExtractor{"Token", "Authorization"}
	httpHeader := http.Header{
		"Authorization": []string{"Bearer xyz456"},
		"Token":         []string{"abc123"},
	}
	token, err := headerExtractor.ExtractHeader(httpHeader)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)
}

// Test BearerExtractor ExtractHeader
func TestBearerExtractor_ExtractHeader(t *testing.T) {
	bearerExtractor := BearerExtractor{}
	httpHeader := http.Header{
		"Authorization": []string{"Bearer xyz456"},
	}
	token, err := bearerExtractor.ExtractHeader(httpHeader)
	assert.Equal(t, err, nil)
	assert.Equal(t, "xyz456", token)
}

// Test AuthorizationExtractor ExtractHeader
func TestAuthorizationExtractor_ExtractHeader(t *testing.T) {
	authorizationExtractor := AuthorizationExtractor{}
	httpHeader := http.Header{
		"Authorization": []string{"Bearer xyz456"},
	}
	token, err := authorizationExtractor.ExtractHeader(httpHeader)
	assert.Equal(t, err, nil)
	assert.Equal(t, "Bearer xyz456", token)
}

// Test TokenExtractor ExtractHeader
func TestTokenExtractor_ExtractHeader(t *testing.T) {
	tokenExtractor := TokenExtractor{}
	httpHeader := http.Header{
		"Token": []string{"abc123"},
	}
	token, err := tokenExtractor.ExtractHeader(httpHeader)
	assert.Equal(t, err, nil)
	assert.Equal(t, "abc123", token)
}

func TestHeaderExtractor_ExtractQuery(t *testing.T) {
	type args struct {
		query url.Values
	}
	tests := []struct {
		name    string
		h       HeaderExtractor
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestHeaderExtractor_ExtractQuery_1",
			h:    HeaderExtractor{"Token", "Authorization"},
			args: args{
				query: url.Values{
					"Token": []string{"abc123"},
				},
			},
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "TestHeaderExtractor_ExtractQuery_2",
			h:    HeaderExtractor{"Token", "Authorization"},
			args: args{
				query: url.Values{
					"Token":         []string{"abc123"},
					"Authorization": []string{"Bearer xyz456"},
				},
			},
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "TestHeaderExtractor_ExtractQuery_3",
			h:    HeaderExtractor{"Token", "Authorization"},
			args: args{
				query: url.Values{
					"X-Header": []string{"abc123"},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.h.ExtractQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("HeaderExtractor.ExtractQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HeaderExtractor.ExtractQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBearerExtractor_ExtractQuery(t *testing.T) {
	type args struct {
		query url.Values
	}
	tests := []struct {
		name    string
		e       BearerExtractor
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "TestBearerExtractor_ExtractQuery_1",
			e:    BearerExtractor{},
			args: args{
				query: url.Values{
					"Token": []string{"abc123"},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "TestBearerExtractor_ExtractQuery_2",
			e:    BearerExtractor{},
			args: args{
				query: url.Values{
					"Authorization": []string{"Bearer xyz456"},
				},
			},
			want:    "xyz456",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := BearerExtractor{}
			got, err := e.ExtractQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("BearerExtractor.ExtractQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BearerExtractor.ExtractQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorizationExtractor_ExtractQuery(t *testing.T) {
	type args struct {
		query url.Values
	}
	tests := []struct {
		name    string
		e       AuthorizationExtractor
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestAuthorizationExtractor_ExtractQuery_1",
			e:    AuthorizationExtractor{},
			args: args{
				query: url.Values{
					"Authorization": []string{"Bearer xyz456"},
				},
			},
			want:    "Bearer xyz456",
			wantErr: false,
		},
		{
			name: "TestAuthorizationExtractor_ExtractQuery_2",
			e:    AuthorizationExtractor{},
			args: args{
				query: url.Values{
					"Token": []string{"abc123"},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := AuthorizationExtractor{}
			got, err := e.ExtractQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthorizationExtractor.ExtractQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AuthorizationExtractor.ExtractQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTokenExtractor_ExtractQuery(t *testing.T) {
	type args struct {
		query url.Values
	}
	tests := []struct {
		name    string
		e       TokenExtractor
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestTokenExtractor_ExtractQuery_1",
			e:    TokenExtractor{},
			args: args{
				query: url.Values{
					"Token": []string{"abc123"},
				},
			},
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "TestTokenExtractor_ExtractQuery_2",
			e:    TokenExtractor{},
			args: args{
				query: url.Values{
					"Authorization": []string{"Bearer xyz456"},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := TokenExtractor{}
			got, err := e.ExtractQuery(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("TokenExtractor.ExtractQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TokenExtractor.ExtractQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
