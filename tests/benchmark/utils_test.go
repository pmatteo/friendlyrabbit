package benchmark_tests

import (
	"fmt"
	"testing"

	"github.com/pmatteo/friendlyrabbit/internal/utils/crypto"
)

func BenchmarkGetHashWithArgon2(b *testing.B) {
	password := "SuperStreetFighter2Turbo"
	salt := "MBisonDidNothingWrong"

	for i := 0; i < 1000; i++ {
		crypto.GetHashWithArgon(fmt.Sprintf(password+"-%d", i), salt, 1, 12, 64, 32)
	}
}
