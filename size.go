package ring


// The best way to change these is to copy this package
// Modifying it through the linker may affect dependencies
// This is the only file you need to change to adjust the ring size
//
// NOTE(as): Some tests will not run if the Size is > 256
const (
	// Size is the number of elements in the ring
	// this must be a power of 2
	Size         = 256

	// CacheLine is the size of the system L1 cache
	// If your system prefetches change it to 128
	// You can check this in the BIOS/EFI settings
	CacheLine    = 64
)
