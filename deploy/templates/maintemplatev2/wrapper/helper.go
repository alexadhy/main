package wrapper

import (
	"context"
	"fmt"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/certutils"
	"time"

	"go.amplifyedge.org/sys-share-v2/sys-core/service/clihelper"
	"go.amplifyedge.org/sys-share-v2/sys-core/service/logging"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	bscfg "go.amplifyedge.org/main-v2/deploy/bootstrapper/service/go"
	bsPkg "go.amplifyedge.org/main-v2/deploy/bootstrapper/service/go/pkg"
	discoSvc "go.amplifyedge.org/mod-v2/mod-disco/service/go"
	discoPkg "go.amplifyedge.org/mod-v2/mod-disco/service/go/pkg"
	discoRpc "go.amplifyedge.org/mod-v2/mod-disco/service/go/rpc/v2"
	sysSharePkg "go.amplifyedge.org/sys-share-v2/sys-account/service/go/pkg"
	sysCorePkg "go.amplifyedge.org/sys-share-v2/sys-core/service/go/pkg"
	corebus "go.amplifyedge.org/sys-share-v2/sys-core/service/go/pkg/bus"
	sysPkg "go.amplifyedge.org/sys-v2/main/pkg"
	"go.amplifyedge.org/sys-v2/sys-core/service/go/pkg/coredb"
)

const (
	defaultTimeout      = 5 * time.Second
	errCreateSysService = "error while creating sys-%s service: %v"
	errCreateModService = "error while creating mod-%s service: %v"
)

type MainService struct {
	Sys   *sysPkg.SysServices
	Disco *discoPkg.ModDiscoService
	BS    *bsPkg.BootstrapService
}

type MainSdkCli struct {
	SysAccount   *sysSharePkg.SysAccountProxyClient
	SysCore      *sysCorePkg.SysCoreProxyClient
	SysMail      *sysCorePkg.SysMailProxyClient
	SysBus       *sysCorePkg.SysBusProxyClient
	ModDisco     *cobra.Command
	BootstrapCLI *cobra.Command
}

func NewMainService(
	logger logging.Logger,
	sscfg *sysPkg.SysServiceConfig,
	cbus *corebus.CoreBus,
	mainCfg *MainServerConfig,
	discocfg *discoSvc.ModDiscoConfig,
	bsconfig *bscfg.BootstrapConfig,
) (*MainService, error) {
	discodb, err := coredb.NewCoreDB(logger, &discocfg.SysCoreConfig, nil)
	if err != nil {
		logger.Fatalf("error creating mod-disco database: %v", err)
	}
	var clientConn *grpc.ClientConn
	var dialOpts grpc.DialOption
	// cli options
	hostPort := fmt.Sprintf("%s:%d", mainCfg.HostAddress, mainCfg.Port)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	if mainCfg.TLS.Enable {
		creds, err := certutils.ClientLoadCA(mainCfg.TLS.RootCAPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load CA Root path: %v", err)
		}
		dialOpts = grpc.WithTransportCredentials(creds)
	} else {
		dialOpts = grpc.WithInsecure()
	}
	clientConn, err = grpc.DialContext(
		ctx,
		hostPort,
		dialOpts,
	)

	if err != nil {
		return nil, fmt.Errorf("unable to create new client conn: %v", err)
	}
	// initiate all sys-* service
	sysSvc, err := sysPkg.NewService(sscfg, mainCfg.Domain)
	if err != nil {
		return nil, fmt.Errorf(errCreateSysService, "all", err)
	}

	discoSvcCfg, err := discoPkg.NewModDiscoServiceConfig(
		logger, discodb, discocfg, cbus, clientConn,
		// logger, discodb, discocfg, cbus, nil,
	)
	if err != nil {
		return nil, fmt.Errorf(errCreateModService, "disco", err)
	}
	modDiscoSvc, err := discoPkg.NewModDiscoService(discoSvcCfg, sysSvc.SysAccountSvc.AllDBs)
	if err != nil {
		return nil, fmt.Errorf(errCreateModService, "disco", err)
	}

	bsService := bsPkg.NewBootstrapService(
		bsconfig,
		logger,
		sysSvc.SysAccountSvc.AuthRepo,
		modDiscoSvc.ModDiscoRepo,
		// clientConn,
		nil,
		cbus,
	)

	ms := &MainService{
		Sys:   sysSvc,
		Disco: modDiscoSvc,
		BS:    bsService,
	}
	return ms, nil
}

func NewMainCLI(
	logger logging.Logger,
	mainCfg *MainClientConfig,
	bsconfig *bscfg.BootstrapConfig,
	cbus *corebus.CoreBus,
) (*MainSdkCli, error) {
	var clientConn *grpc.ClientConn
	dialOpts := []grpc.DialOption{grpc.WithBlock()}
	// cli options
	var cliOpts *clihelper.CLIOptions

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	hostPort := fmt.Sprintf("%s:%d", mainCfg.Domain, mainCfg.Port)

	if !mainCfg.IsLocal && mainCfg.TLS.Enable {
		err := certutils.DownloadCACert(mainCfg.TLS.RootCAPath, mainCfg.Domain)
		creds, err := certutils.ClientLoadCA(mainCfg.TLS.RootCAPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load CA from domain: %s => %v", mainCfg.Domain, err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		cliOpts = clihelper.CLIWrapper(mainCfg.TLS.RootCAPath, hostPort, ".env")
	} else if mainCfg.TLS.Enable && mainCfg.IsLocal {
		logger.Infof("loading certificate from local: %v", mainCfg.TLS.RootCAPath)
		creds, err := certutils.ClientLoadCA(mainCfg.TLS.RootCAPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load CA Root path: %v", err)
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		cliOpts = clihelper.CLIWrapper(mainCfg.TLS.RootCAPath, hostPort, ".env")
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
		cliOpts = clihelper.CLIWrapper("", hostPort, ".env")
	}
	clientOptions := cliOpts.GetAllOptions()
	cliOpts.RegisterAuthDialer()
	var err error
	clientConn, err = grpc.DialContext(
		ctx,
		hostPort,
		dialOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new client conn: %v", err)
	}

	bsService := bsPkg.NewBootstrapService(
		bsconfig,
		logger,
		nil,
		nil,
		clientConn,
		cbus,
	)
	sysSharePkg.NewSysShareProxyClient(clientOptions...)

	bootstrapCli := bsPkg.NewBootstrapCLI(bsService.BsRepo, clientOptions...)
	mcli := &MainSdkCli{
		SysAccount:   sysSharePkg.NewSysShareProxyClient(clientOptions...),
		SysCore:      sysCorePkg.NewSysCoreProxyClient(clientOptions...),
		SysMail:      sysCorePkg.NewSysMailProxyClient(clientOptions...),
		SysBus:       sysCorePkg.NewSysBusProxyClient(clientOptions...),
		ModDisco:     discoRpc.SurveyServiceClientCommand(clientOptions...),
		BootstrapCLI: bootstrapCli,
	}
	return mcli, nil
}
