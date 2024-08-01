package authorize

import (
	"context"
	"testing"

	"github.com/r3labs/diff/v3"
)

func TestNewInnerAuthorize(t *testing.T) {
	type args struct {
		option *InnerAuthorizeConfig
	}
	tests := []struct {
		name      string
		args      args
		want      *InnerAuthorize
		wantPanic bool
	}{
		// TODO: Add test cases.
		{
			name: "TestNewInnerAuthorize",
			args: args{
				option: &InnerAuthorizeConfig{
					Secret: "test",
				},
			},
			want: &InnerAuthorize{},
		},
		{
			name: "TestNewInnerAuthorize1",
			args: args{
				option: &InnerAuthorizeConfig{
					Secret: "test1",
				},
			},
			want: &InnerAuthorize{},
		},
		{
			name: "TestNewInnerAuthoriz2",
			args: args{
				option: &InnerAuthorizeConfig{
					Secret: "*&@^!&#$*$@#*!(SD~AD><?)",
				},
			},
			want: &InnerAuthorize{},
		},
		{
			name: "TestNewInnerAuthorize3",
			args: args{
				option: &InnerAuthorizeConfig{
					Secret: "test3",
				},
			},
			want: &InnerAuthorize{},
		},
		{
			name: "TestNewInnerAuthorize4",
			args: args{
				option: &InnerAuthorizeConfig{
					Secret: "",
				},
			},
			want:      &InnerAuthorize{},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.wantPanic {
						return
					}
					t.Errorf("name %s NewInnerAuthorize() recover = %v", tt.name, r)
				}
			}()
			NewInnerAuthorize(tt.args.option)
		})
	}
}

func TestInnerAuthorize_GenerateToken(t *testing.T) {
	type args struct {
		ctx  context.Context
		user IAuthorizeOther
	}
	tests := []struct {
		name           string
		innerAuthorize *InnerAuthorize
		args           args
		want           *UserAuthorizeOther
		wantErr        bool
	}{
		// TODO: Add test cases.
		{
			name: "TestInner1",
			innerAuthorize: NewInnerAuthorize(&InnerAuthorizeConfig{
				Secret: "*&@^!&#$*$@#*!(SD~AD><?)",
			}),
			args: args{
				ctx: context.Background(),
				user: &UserAuthorizeOther{
					Id:   "app_1",
					Type: "service_type",
					Name: "service_name",
				},
			},
			want: &UserAuthorizeOther{
				Id:   "app_1",
				Type: "service_type",
				Name: "service_name",
			},
		},
		{
			name: "TestInner2",
			innerAuthorize: NewInnerAuthorize(&InnerAuthorizeConfig{
				Secret: "*&@^!&#$*$@#*!(SD~AD><?)",
			}),
			args: args{
				ctx: context.Background(),
				user: &UserAuthorizeOther{
					Id:   "app_1",
					Type: "service_type",
					Name: "service_name",
				},
			},
			want: &UserAuthorizeOther{
				Id:   "app_2",
				Type: "service_type2",
				Name: "service_name2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, err := tt.innerAuthorize.GenerateToken(tt.args.ctx, tt.args.user)
			if err != nil {
				t.Errorf("InnerAuthorize.GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log(gotStr)
			_, err = tt.innerAuthorize.VerifyToken(tt.args.ctx, gotStr, tt.args.user)
			if err != nil {
				t.Errorf("InnerAuthorize.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			c, err := diff.Diff(tt.args.user, tt.want)
			if err != nil {
				t.Errorf("Diff error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(c) != 0) != tt.wantErr {
				t.Errorf("Diff = %v, want %v", c, tt.want)
			}
		})
	}
}

func TestInnerAuthorize_VerifyToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token string
		user  IAuthorizeOther
	}
	tests := []struct {
		name           string
		innerAuthorize *InnerAuthorize
		args           args
		want           *UserAuthorizeOther
		wantErr        bool
	}{
		// TODO: Add test cases.
		{
			name: "TestInner1",
			innerAuthorize: NewInnerAuthorize(&InnerAuthorizeConfig{
				Secret: "*&@^!&#$*$@#*!(SD~AD><?)",
			}),
			args: args{
				ctx:   context.Background(),
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpbm5lckF1dGhvcml6ZUluZm8iOiJleUpoYkdjaU9pSkJNVEk0UzFjaUxDSmxibU1pT2lKQk1qVTJSME5OSW4wLmJ1azVmOTVYTjVJQjV6VnZBdk03YzVhTDhlNThmQkJYUk1VYjhOblo3ajZCamZMZ2hBRXVFQS5MVUVlMlctZmZTNlpZN0ZoLl9OQkNLcFgxcFFlYjJEekNvS3BGd0RwdTJrVW50aElSdDJXcHpyaG45SVRrZGdHdW5YSVE1dFRXN1djdjFmeEc0cnBzbzNBR2NXc2owWTFKUmlac2FNMVZEM0loX1EuSmxqYkRacWlYNVM2QlNFRU5PTE5yUSIsImlzcyI6ImlubmVyQXV0aG9yaXplSW5mbyJ9.hjr6EzljLJvnSLb1HjWMZ7rQufjHn4gE8GJGwcnXT8g",
				user:  &UserAuthorizeOther{},
			},
			want: &UserAuthorizeOther{
				Id:   "app_1",
				Type: "service_type",
				Name: "service_name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.innerAuthorize.VerifyToken(tt.args.ctx, tt.args.token, tt.args.user)
			if err != nil {
				t.Errorf("InnerAuthorize.VerifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// t.Logf("tt.args.user %v", tt.args.user)
			// return
			c, err := diff.Diff(tt.args.user, tt.want)
			if err != nil {
				t.Errorf("Diff error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (len(c) != 0) != tt.wantErr {
				t.Errorf("Diff = %v, want %v", c, tt.want)
			}
		})
	}
}
