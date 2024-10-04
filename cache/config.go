package cache

// A Config represents an ActiveCache parameters configuration
type Config struct {
	// CleanerInterval is the interval in ms that cleaner will run
	//
	// If value is less than `MinCleanerInterval` then `DefaultCleanerInterval` will be set
	CleanerInterval int

	// KeysAmountByCycle is the amount of keys that will be checked
	//
	// If value is less than `MinKeysAmountByCycle` then `DefaultKeysAmountByCycle` will be set
	KeysAmountByCycle int
}

// DefaultConfig returns a Config pointer instance
//
// with default values for parameters
func DefaultConfig() *Config {
	return &Config{
		CleanerInterval:   DefaultCleanerInterval,
		KeysAmountByCycle: DefaultKeysAmountByCycle,
	}
}
