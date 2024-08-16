package baseapp

import (
	"context"

	"github.com/berachain/offchain-sdk/client/eth"
	"github.com/berachain/offchain-sdk/job"
	"github.com/berachain/offchain-sdk/log"
	"github.com/berachain/offchain-sdk/server"
	"github.com/berachain/offchain-sdk/telemetry"

	ethdb "github.com/ethereum/go-ethereum/ethdb"
)

// BaseApp represents the base application.
type BaseApp struct {
	// name is the name of the application.
	name string

	// logger is the logger for the baseapp.
	logger log.Logger

	// jobMgr is the job manager for handling jobs.
	jobMgr *JobManager

	// svr is the server for the baseapp.
	svr *server.Server
}

// New creates a new BaseApp instance.
func New(
	name string,
	logger log.Logger,
	ethClient eth.Client,
	jobs []job.Basic,
	db ethdb.KeyValueStore,
	svr *server.Server,
	metrics telemetry.Metrics,
) *BaseApp {
	return &BaseApp{
		name:   name,
		logger: logger,
		jobMgr: NewManager(
			jobs,
			&contextFactory{
				connPool: ethClient,
				logger:   logger,
				db:       db,
				metrics:  metrics,
			},
		),
		svr: svr,
	}
}

// Logger returns the logger for the baseapp with a namespace.
func (b *BaseApp) Logger() log.Logger {
	return b.logger.With("namespace", "baseapp")
}

// Start initializes and starts the baseapp components.
func (b *BaseApp) Start(ctx context.Context) error {
	b.Logger().Info("Attempting to start")
	defer b.Logger().Info("Successfully started")

	// Start the job manager and the producers.
	b.jobMgr.Start(ctx)
	b.jobMgr.RunProducers(ctx)

	if b.svr == nil {
		b.Logger().Info("No HTTP server registered, skipping")
	} else {
		go func() {
			if err := b.svr.Start(ctx); err != nil {
				b.Logger().Error("HTTP server failed to start", "error", err)
			}
		}()
	}

	return nil
}

// Stop shuts down the baseapp and its components.
func (b *BaseApp) Stop() {
	b.Logger().Info("Attempting to stop")
	defer b.Logger().Info("Successfully stopped")

	b.jobMgr.Stop()
	if b.svr != nil {
		b.svr.Stop()
	}
}
