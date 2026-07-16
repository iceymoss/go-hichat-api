package config

import "testing"

func Test_Config_ApplyEnvironment_OverridesPublicAndTURN(t *testing.T) {
	t.Setenv("PUBLIC_IP", "203.0.113.10")
	t.Setenv("SFU_UDP_MIN_PORT", "50000")
	t.Setenv("SFU_UDP_MAX_PORT", "50200")
	t.Setenv("TURN_URLS", "turn:203.0.113.10:3478?transport=udp, turn:203.0.113.10:3478?transport=tcp")
	t.Setenv("TURN_SECRET", "test-secret")
	t.Setenv("TURN_TTL_SECONDS", "7200")

	var c Config
	if err := c.ApplyEnvironment(); err != nil {
		t.Fatalf("ApplyEnvironment: %v", err)
	}

	if c.Public.IP != "203.0.113.10" || c.Public.UDPPortMin != 50000 || c.Public.UDPPortMax != 50200 {
		t.Fatalf("public config = %+v", c.Public)
	}
	if len(c.Turn.URLs) != 2 || c.Turn.Secret != "test-secret" || c.Turn.TTLSeconds != 7200 {
		t.Fatalf("turn config = %+v", c.Turn)
	}
}

func Test_Config_ApplyEnvironment_RejectsInvalidPortRange(t *testing.T) {
	t.Setenv("SFU_UDP_MIN_PORT", "50200")
	t.Setenv("SFU_UDP_MAX_PORT", "50000")

	var c Config
	if err := c.ApplyEnvironment(); err == nil {
		t.Fatal("ApplyEnvironment accepted an invalid UDP port range")
	}
}
