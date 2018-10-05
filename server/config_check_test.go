// Copyright 2018 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestConfigCheck(t *testing.T) {
	tests := []struct {
		// name is the name of the test.
		name string

		// config is content of the configuration file.
		config string

		// defaultErr is the error we get pedantic checks are not enabled.
		defaultErr error

		// pedanticErr is the error we get when pedantic checks are enabled.
		pedanticErr error

		// errorLine is the location of the error.
		errorLine int

		// errorPos is the position of the error.
		errorPos int

		// warning errors also include a reason optionally.
		reason string

		// newDefaultErr is a configuration error that includes source of error.
		newDefaultErr error
	}{
		{
			name: "when unknown field is used at top level",
			config: `
                monitor = "127.0.0.1:4442"
                `,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "monitor"`),
			errorLine:   2,
			errorPos:    17,
		},
		{
			name: "when default permissions are used at top level",
			config: `
                "default_permissions" {
                  publish = ["_SANDBOX.>"]
                  subscribe = ["_SANDBOX.>"]
                }
                `,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "default_permissions"`),
			errorLine:   2,
			errorPos:    18,
		},
		{
			name: "when authorization config is empty",
			config: `
		authorization = {
		}
		`,
			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when authorization config has unknown fields",
			config: `
		authorization = {
		  foo = "bar"
		}
		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "foo"`),
			errorLine:   3,
			errorPos:    5,
		},
		{
			name: "when authorization config has unknown fields",
			config: `
		port = 4222

		authorization = {
		  user = "hello"
		  foo = "bar"
		  password = "world"
		}

		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "foo"`),
			errorLine:   6,
			errorPos:    5,
		},
		{
			name: "when user authorization config has unknown fields",
			config: `
		authorization = {
		  users = [
		    {
		      user = "foo"
		      pass = "bar"
		      token = "quux"
		    }
		  ]
		}
		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "token"`),
			errorLine:   7,
			errorPos:    9,
		},
		{
			name: "when user authorization permissions config has unknown fields",
			config: `
		authorization {
		  permissions {
		    subscribe = {}
		    inboxes = {}
		    publish = {}
		  }
		}
		`,
			defaultErr:  errors.New(`Unknown field inboxes parsing permissions`),
			pedanticErr: errors.New(`unknown field "inboxes"`),
			errorLine:   5,
			errorPos:    7,
		},
		{
			name: "when user authorization permissions config has unknown fields within allow or deny",
			config: `
		authorization {
		  permissions {
		    subscribe = {
		      allow = ["hello", "world"]
		      deny = ["foo", "bar"]
		      denied = "_INBOX.>"
		    }
		    publish = {}
		  }
		}
		`,
			defaultErr:  errors.New(`Unknown field name "denied" parsing subject permissions, only 'allow' or 'deny' are permitted`),
			pedanticErr: errors.New(`unknown field "denied"`),
			errorLine:   7,
			errorPos:    9,
		},
		{
			name: "when user authorization permissions config has unknown fields within allow or deny",
			config: `
		authorization {
		  permissions {
		    publish = {
		      allow = ["hello", "world"]
		      deny = ["foo", "bar"]
		      allowed = "_INBOX.>"
		    }
		    subscribe = {}
		  }
		}
		`,
			defaultErr:  errors.New(`Unknown field name "allowed" parsing subject permissions, only 'allow' or 'deny' are permitted`),
			pedanticErr: errors.New(`unknown field "allowed"`),
			errorLine:   7,
			errorPos:    9,
		},
		{
			name: "when user authorization permissions config has unknown fields using arrays",
			config: `
		authorization {

		 default_permissions {
		   subscribe = ["a"]
		   publish = ["b"]
		   inboxes = ["c"]
		 }

		 users = [
		   {
		     user = "foo"
		     pass = "bar"
		   }
		  ]
		}
		`,
			defaultErr:  errors.New(`Unknown field inboxes parsing permissions`),
			pedanticErr: errors.New(`unknown field "inboxes"`),
			errorLine:   7,
			errorPos:    6,
		},
		{
			name: "when user authorization permissions config has unknown fields using strings",
			config: `
		authorization {

		 default_permissions {
		   subscribe = "a"
		   requests = "b"
		   publish = "c"
		 }

		 users = [
		   {
		     user = "foo"
		     pass = "bar"
		   }
		  ]
		}
		`,
			defaultErr:  errors.New(`Unknown field requests parsing permissions`),
			pedanticErr: errors.New(`unknown field "requests"`),
			errorLine:   6,
			errorPos:    6,
		},
		{
			name: "when user authorization permissions config is empty",
			config: `
		authorization = {
		  users = [
		    {
		      user = "foo", pass = "bar", permissions = {
		      }
		    }
		  ]
		}
		`,
			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when unknown permissions are included in config",
			config: `
		authorization = {
		  users = [
		    {
		      user = "foo", pass = "bar", permissions {
		        inboxes = true
		      }
		    }
		  ]
		}
		`,
			defaultErr:  errors.New(`Unknown field inboxes parsing permissions`),
			pedanticErr: errors.New(`unknown field "inboxes"`),
			errorLine:   6,
			errorPos:    11,
		},
		{
			name: "when clustering config is empty",
			config: `
		cluster = {
		}
		`,

			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when unknown option is in clustering config",
			config: `
		# NATS Server Configuration
		port = 4222

		cluster = {

		  port = 6222

		  foo = "bar"

		  authorization {
		    user = "hello"
		    pass = "world"
		  }

		}
		`,

			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "foo"`),
			errorLine:   9,
			errorPos:    5,
		},
		{
			name: "when unknown option is in clustering authorization config",
			config: `
		cluster = {
		  authorization {
		    foo = "bar"
		  }
		}
		`,

			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "foo"`),
			errorLine:   4,
			errorPos:    7,
		},
		{
			name: "when unknown option is in clustering authorization permissions config",
			config: `
		cluster = {
		  authorization {
		    user = "foo"
		    pass = "bar"
		    permissions = {
		      hello = "world"
		    }
		  }
		}
		`,
			defaultErr:  errors.New(`Unknown field hello parsing permissions`),
			pedanticErr: errors.New(`unknown field "hello"`),
			errorLine:   7,
			errorPos:    9,
		},
		{
			name: "when unknown option is in tls config",
			config: `
		tls = {
		  hello = "world"
		}
		`,
			defaultErr:  errors.New(`error parsing tls config, unknown field ["hello"]`),
			pedanticErr: errors.New(`unknown field "hello"`),
			errorLine:   3,
			errorPos:    5,
		},
		{
			name: "when unknown option is in cluster tls config",
			config: `
		cluster {
		  tls = {
		    foo = "bar"
		  }
		}
		`,
			// Backwards compatibility: also report error by default even if pedantic checks disabled.
			defaultErr:  errors.New(`error parsing tls config, unknown field ["foo"]`),
			pedanticErr: errors.New(`unknown field "foo"`),
			errorLine:   4,
			errorPos:    7,
		},
		{
			name: "when using cipher suites in the TLS config",
			config: `
		tls = {
		    cipher_suites: [
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
		    ]
		    preferences = []
		}
		`,
			defaultErr:  errors.New(`error parsing tls config, unknown field ["preferences"]`),
			pedanticErr: errors.New(`unknown field "preferences"`),
			errorLine:   7,
			errorPos:    7,
		},
		{
			name: "when using curve preferences in the TLS config",
			config: `
		tls = {
		    curve_preferences: [
			"CurveP256",
			"CurveP384",
			"CurveP521"
		    ]
		    suites = []
		}
		`,
			defaultErr:  errors.New(`error parsing tls config, unknown field ["suites"]`),
			pedanticErr: errors.New(`unknown field "suites"`),
			errorLine:   8,
			errorPos:    7,
		},
		{
			name: "when unknown option is in cluster config with defined routes",
			config: `
		cluster {
		  port = 6222
		  routes = [
		    nats://127.0.0.1:6222
		  ]
		  peers = []
		}
		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "peers"`),
			errorLine:   7,
			errorPos:    5,
		},
		{
			name: "when used as variable in authorization block it should not be considered as unknown field",
			config: `
		# listen:   127.0.0.1:-1
		listen:   127.0.0.1:4222

		authorization {
		  # Superuser can do anything.
		  super_user = {
		    publish = ">"
		    subscribe = ">"
		  }

		  # Can do requests on foo or bar, and subscribe to anything
		  # that is a response to an _INBOX.
		  #
		  # Notice that authorization filters can be singletons or arrays.
		  req_pub_user = {
		    publish = ["req.foo", "req.bar"]
		    subscribe = "_INBOX.>"
		  }

		  # Setup a default user that can subscribe to anything, but has
		  # no publish capabilities.
		  default_user = {
		    subscribe = "PUBLIC.>"
		  }

		  unused = "hello"

		  # Default permissions if none presented. e.g. susan below.
		  default_permissions: $default_user

		  # Users listed with persmissions.
		  users = [
		    {user: alice, password: foo, permissions: $super_user}
		    {user: bob,   password: bar, permissions: $req_pub_user}
		    {user: susan, password: baz}
		  ]
		}
		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`unknown field "unused"`),
			errorLine:   27,
			errorPos:    5,
		},
		{
			name: "when used as variable in top level config it should not be considered as unknown field",
			config: `
		monitoring_port = 8222

		http_port = $monitoring_port

		port = 4222
		`,
			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when used as variable in cluster config it should not be considered as unknown field",
			config: `
		cluster {
		  clustering_port = 6222
		  port = $clustering_port
		}
		`,
			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when setting permissions within cluster authorization block",
			config: `
		cluster {
		  authorization {
		    permissions = {
		      publish = { allow = ["foo", "bar"] }
		    }
		  }

		  permissions = {
		    publish = { deny = ["foo", "bar"] }
		  }
		}
		`,
			defaultErr:  nil,
			pedanticErr: errors.New(`invalid use of field "authorization"`),
			errorLine:   3,
			errorPos:    5,
			reason:      `setting "permissions" within cluster authorization block is deprecated`,
		},
		/////////////////////
		// ACCOUNTS	   //
		/////////////////////
		{
			name: "when accounts block is correctly configured",
			config: `
		http_port = 8222

		accounts {

		  #
		  # synadia > nats.io, cncf
		  #
		  synadia {
		    # SAADJL5XAEM6BDYSWDTGVILJVY54CQXZM5ZLG4FRUAKB62HWRTPNSGXOHA
		    nkey = "AC5GRL36RQV7MJ2GT6WQSCKDKJKYTK4T2LGLWJ2SEJKRDHFOQQWGGFQL"

		    users [
		      {
		        # SUAEL6RU3BSDAFKOHNTEOK5Q6FTM5FTAMWVIKBET6FHPO4JRII3CYELVNM
		        nkey = "UCARKS2E3KVB7YORL2DG34XLT7PUCOL2SVM7YXV6ETHLW6Z46UUJ2VZ3"
		      }
		    ]

		    exports = [
		      { service: "synadia.requests", accounts: [nats, cncf] }
		    ]
		  }

		  #
		  # nats < synadia
		  #
		  nats {
		    # SUAJTM55JH4BNYDA22DMDZJSRBRKVDGSLYK2HDIOCM3LPWCDXIDV5Q4CIE
		    nkey = "ADRZ42QBM7SXQDXXTSVWT2WLLFYOQGAFC4TO6WOAXHEKQHIXR4HFYJDS"

		    users [
		      {
		        # SUADZTYQAKTY5NQM7XRB5XR3C24M6ROGZLBZ6P5HJJSSOFUGC5YXOOECOM
		        nkey = "UD6AYQSOIN2IN5OGC6VQZCR4H3UFMIOXSW6NNS6N53CLJA4PB56CEJJI"
		      }
		    ]

		    imports = [
		      # This account has to send requests to 'nats.requests' subject
		      { service: { account: "synadia", subject: "synadia.requests" }, to: "nats.requests" }
		    ]
		  }

		  #
		  # cncf < synadia
		  #
		  cncf {
		    # SAAFHDZX7SGZ2SWHPS22JRPPK5WX44NPLNXQHR5C5RIF6QRI3U65VFY6C4
		    nkey = "AD4YRVUJF2KASKPGRMNXTYKIYSCB3IHHB4Y2ME6B2PDIV5QJ23C2ZRIT"

		    users [
		      {
		        # SUAKINP3Z2BPUXWOFSW2FZC7TFJCMMU7DHKP2C62IJQUDASOCDSTDTRMJQ
		        nkey = "UB57IEMPG4KOTPFV5A66QKE2HZ3XBXFHVRCCVMJEWKECMVN2HSH3VTSJ"
		      }
		    ]

		    imports = [
		      # This account has to send requests to 'synadia.requests' subject
		      { service: { account: "synadia", subject: "synadia.requests" } }
		    ]
		  }
		}
				`,
			defaultErr:  nil,
			pedanticErr: nil,
		},
		{
			name: "when accounts block has unknown fields",
			config: `
		http_port = 8222

		accounts {
                  foo = "bar"
		}
				`,
			newDefaultErr: errors.New(`Expected map entries for accounts`),
			errorLine:     4,
			errorPos:      3,
		},
		{
			name: "when accounts block defines a global account",
			config: `
		http_port = 8222

		accounts {
                  $G = {
                  }
		}
				`,
			newDefaultErr: errors.New(`"$G" is a Reserved Account`),
			errorLine:     4,
			errorPos:      3,
		},
		{
			name: "when accounts block uses an invalid public key",
			config: `
		accounts {
                  synadia = {
                    nkey = "invalid"
                  }
		}
				`,
			newDefaultErr: errors.New(`Not a valid public nkey for an account: "invalid"`),
			errorLine:     4,
			errorPos:      21,
		},
		{
			name: "when accounts block uses an invalid public key",
			config: `
		accounts {
                  synadia = {
                    nkey = "invalid"
                  }
		}
				`,
			newDefaultErr: errors.New(`Not a valid public nkey for an account: "invalid"`),
			errorLine:     4,
			errorPos:      21,
		},
	}

	checkConfig := func(config string, pedantic bool) error {
		opts := &Options{
			CheckConfig: pedantic,
		}
		return opts.ProcessConfigFile(config)
	}

	checkErr := func(t *testing.T, err, expectedErr error) {
		t.Helper()
		switch {
		case err == nil && expectedErr == nil:
			// OK
		case err != nil && expectedErr == nil:
			t.Errorf("Unexpected error after processing config: %s", err)
		case err == nil && expectedErr != nil:
			t.Errorf("Expected %q error after processing invalid config but got nothing", expectedErr)
		}
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf := createConfFile(t, []byte(test.config))
			defer os.Remove(conf)

			t.Run("with pedantic check enabled", func(t *testing.T) {
				err := checkConfig(conf, true)
				var expectedErr error

				// New default errors also include source of error
				// like an error reported when running with pedantic flag.
				if test.newDefaultErr != nil {
					expectedErr = test.newDefaultErr
				} else if test.pedanticErr != nil {
					expectedErr = test.pedanticErr
				}

				if err != nil && expectedErr != nil {
					msg := fmt.Sprintf("%s:%d:%d: %s", conf, test.errorLine, test.errorPos, expectedErr.Error())
					if test.reason != "" {
						msg += ": " + test.reason
					}
					if err.Error() != msg {
						t.Errorf("Expected %q, got %q", msg, err.Error())
					}
				}

				checkErr(t, err, expectedErr)
			})

			t.Run("with pedantic check disabled", func(t *testing.T) {
				err := checkConfig(conf, false)

				// Gradually move to all errors including source of the error.
				if err != nil && test.newDefaultErr != nil {
					expectedErr := test.newDefaultErr
					source := fmt.Sprintf("%s:%d:%d", conf, test.errorLine, test.errorPos)
					expectedMsg := fmt.Sprintf("%s: %s", source, expectedErr.Error())
					if err.Error() != expectedMsg {
						t.Errorf("\nExpected: \n%q, \ngot: \n%q", expectedMsg, err.Error())
					}
				} else if err != nil && test.defaultErr != nil {
					expectedErr := test.defaultErr
					if err != nil && expectedErr != nil && err.Error() != expectedErr.Error() {
						t.Errorf("Expected: \n%q, \ngot: \n%q", expectedErr.Error(), err.Error())
					}
					checkErr(t, err, test.defaultErr)
				}
			})
		})
	}
}

func TestConfigCheckIncludes(t *testing.T) {
	// Check happy path first using pedantic mode.
	opts := &Options{
		CheckConfig: true,
	}
	err := opts.ProcessConfigFile("./configs/include_conf_check_a.conf")
	if err != nil {
		t.Errorf("Unexpected error processing include files with configuration check enabled: %v", err)
	}

	opts = &Options{
		CheckConfig: true,
	}
	err = opts.ProcessConfigFile("./configs/include_bad_conf_check_a.conf")
	if err == nil {
		t.Errorf("Expected error processing include files with configuration check enabled: %v", err)
	}
	expectedErr := errors.New(`configs/include_bad_conf_check_b.conf:10:19: unknown field "monitoring_port"`)
	if err != nil && expectedErr != nil && err.Error() != expectedErr.Error() {
		t.Errorf("Expected: \n%q, got\n: %q", expectedErr.Error(), err.Error())
	}
}
