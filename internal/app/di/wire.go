//go:build wireinject

package di

import (
	"github.bumble.dev/shcherbanich/user-votes-storage/config"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/app"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/app/api"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/messaging/handler"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/application/operation"
	countersRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/counter/repository"
	romancesRepo "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/domain/romance/repository"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/infrastructure/persistence"
	storageV1 "github.bumble.dev/shcherbanich/user-votes-storage/internal/context/voting/interface/api/rest/v1"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/messaging"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/amazon_sns"
	"github.bumble.dev/shcherbanich/user-votes-storage/internal/shared/platform/dynamodb"
	"github.com/google/wire"
)

var PlatformSet = wire.NewSet(
	platform.NewLogger,
)

var ReposSet = wire.NewSet(
	dynamodb.NewDynamoDbClient,
	persistence.NewRomancesRepository,
	persistence.NewCountersRepository,
	wire.Bind(new(romancesRepo.RomancesRepository), new(*persistence.RomancesRepository)),
	wire.Bind(new(countersRepo.CountersRepository), new(*persistence.CountersRepository)),
)

func InitializeApiWebServer(config config.Config) (*app.ApiWebServer, error) {
	wire.Build(
		PlatformSet,
		ReposSet,
		amazon_sns.NewSnsPublisher,
		wire.Bind(new(messaging.Publisher), new(*amazon_sns.SnsPublisher)),
		operation.NewGetRomanceOperation,
		operation.NewDeleteRomanceOperation,
		operation.NewGetUserVoteOperation,
		operation.NewAddUserVoteOperation,
		operation.NewChangeUserVoteOperation,
		operation.NewDeleteUserVoteOperation,
		operation.NewGetLifetimeCountersOperation,
		operation.NewGetHourlyCountersOperation,
		operation.NewDeleteRomancesOperation,
		application.NewVotingService,
		storageV1.NewVotesStorageRoutsRegister,
		api.NewHandlerFactory,
		app.NewApiWebServer,
	)
	return nil, nil
}

func InitializeMessageProcessor(config config.Config) (*app.MessageProcessor, error) {
	wire.Build(
		PlatformSet,
		amazon_sns.NewSnsSubscriber,
		handler.NewDeleteDeleteRomancesHandler,
		wire.Bind(new(messaging.Subscriber), new(*amazon_sns.SnsSubscriber)),
		app.NewMessageProcessor,
	)
	return nil, nil
}
