package main

import (
	"context"
	"fmt"
	"time"

	"github.com/edaniels/golog"
	"go.uber.org/multierr"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"
)

var Model = resource.NewModel("sensehat", "sensor", "lps25h")

const (
	defaultI2Caddr = 0x5c
	// Addresses of lps25h registers.
	lps25hCOMMANDSOFTRESET1 = 0x30
	lps25hCOMMANDSOFTRESET2 = 0xA2
	lps25hCOMMANDPOLLINGH1  = 0x24
	lps25hCOMMANDPOLLINGH2  = 0x00
)

type lps25h struct {
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable
	logger golog.Logger

	bus  board.I2C
	addr byte
}

// Config is used for converting config attributes.
type Config struct {
	Board   string `json:"board"`
	I2CBus  string `json:"i2c_bus"`
	I2cAddr int    `json:"i2c_addr,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (conf *Config) Validate(path string) ([]string, error) {
	var deps []string
	if len(conf.Board) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "board")
	}
	deps = append(deps, conf.Board)
	if len(conf.I2CBus) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "i2c_bus")
	}
	return deps, nil
}

// Readings always returns "hello world".
func (s *lps25h) Readings(ctx context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{"hello": "world"}, nil
}

// We first put our component's constructor in the registry, then tell the module to load it
// Note that all resources must be added before the module is started.
func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *Config]{Constructor: func(
			ctx context.Context,
			deps resource.Dependencies,
			conf resource.Config,
			logger golog.Logger,
		) (sensor.Sensor, error) {
			newConf, err := resource.NativeConfig[*Config](conf)
			if err != nil {
				return nil, err
			}
			return newSensor(ctx, deps, conf.ResourceName(), newConf, logger)
		}})
}

func newSensor(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	conf *Config,
	logger golog.Logger,
) (sensor.Sensor, error) {
	b, err := board.FromDependencies(deps, conf.Board)
	if err != nil {
		return nil, fmt.Errorf("lps25h init: failed to find board: %w", err)
	}
	localB, ok := b.(board.LocalBoard)
	if !ok {
		return nil, fmt.Errorf("board %s is not local", conf.Board)
	}
	i2cbus, ok := localB.I2CByName(conf.I2CBus)
	if !ok {
		return nil, fmt.Errorf("lps25h init: failed to find i2c bus %s", conf.I2CBus)
	}
	addr := conf.I2cAddr
	if addr == 0 {
		addr = defaultI2Caddr
		logger.Warn("using i2c address : 0x44")
	}

	s := &lps25h{
		Named:  name.AsNamed(),
		logger: logger,
		bus:    i2cbus,
		addr:   byte(addr),
	}

	err = s.reset(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// reset will reset the sensor.
func (s *lps25h) reset(ctx context.Context) error {
	handle, err := s.bus.OpenHandle(s.addr)
	if err != nil {
		s.logger.Errorf("can't open lps25h i2c %s", err)
		return err
	}
	err = handle.Write(ctx, []byte{lps25hCOMMANDSOFTRESET1, lps25hCOMMANDSOFTRESET2})
	// wait for chip reset cycle to complete
	time.Sleep(1 * time.Millisecond)
	return multierr.Append(err, handle.Close())
}

func main() {
	utils.ContextualMain(mainWithArgs, golog.NewDevelopmentLogger("LPS25H"))
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	// Instantiate the module itself
	myModule, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	err = myModule.AddModelFromRegistry(ctx, sensor.API, Model)
	if err != nil {
		return err
	}

	// The module is started.
	err = myModule.Start(ctx)
	// Close is deferred and will run automatically when this function returns.
	defer myModule.Close(ctx)
	if err != nil {
		return err
	}

	// This will block (leaving the module running) until the context is cancelled.
	// The utils.ContextualMain catches OS signals and will cancel our context for us when one is sent for shutdown/termination.
	<-ctx.Done()
	// The deferred myModule.Close() will now run.
	return nil
}
