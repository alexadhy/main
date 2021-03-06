package repo

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"go.amplifyedge.org/main-v2/deploy/bootstrapper/service/go/pkg/fakedata"
	bsrpc "go.amplifyedge.org/main-v2/deploy/bootstrapper/service/go/rpc/v2"
	accountRpc "go.amplifyedge.org/sys-share-v2/sys-account/service/go/rpc/v2"
	coreRpc "go.amplifyedge.org/sys-share-v2/sys-core/service/go/rpc/v2"
)

const (
	defaultTimeout = 5 * time.Second
)

func (b *BootstrapRepo) sharedExecutor(ctx context.Context, bsAll *fakedata.BootstrapAll, filename string) (err error) {
	orgs := bsAll.GetOrgs()
	projects := bsAll.GetProjects()
	regs := bsAll.GetRegularUsers()
	if b.accRepo != nil && b.discoRepo != nil {
		err = b.sharedExecv2(ctx, orgs, projects, regs)
	}
	if b.accClient != nil && b.discoClient != nil {
		err = b.sharedExecv3(ctx, orgs, projects, regs)
	}
	if err != nil {
		return err
	}
	return b.setIsActive(filename)
	//return fmt.Errorf("invalid argument, no repo or client defined for bootstrap")
}

func (b *BootstrapRepo) sharedExecv3(ctx context.Context, orgs []*bsrpc.BSOrg, projects []*bsrpc.BSProject, regularUsers []*bsrpc.BSRegularAccount) error {
	var err error
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	err = b.resetAll(ctx)
	if err != nil {
		return err
	}

	for _, org := range orgs {
		if _, err = b.orgClient.NewOrg(ctx, org.GetOrg()); err != nil {
			return err
		}
	}
	for _, proj := range projects {
		if _, err = b.orgClient.NewProject(ctx, proj.GetProject()); err != nil {
			return err
		}
		if _, err = b.discoClient.NewDiscoProject(ctx, proj.GetProjectDetails()); err != nil {
			return err
		}
		if proj.GetSurveySchema() != nil {
			if _, err = b.discoClient.NewSurveyProject(ctx, proj.GetSurveySchema()); err != nil {
				return err
			}
		}
	}

	for _, reg := range regularUsers {
		acc, err := b.accClient.NewAccount(ctx, reg.GetNewAccounts())
		if err != nil {
			return err
		}
		updRequest := &accountRpc.AccountUpdateRequest{
			Id:       acc.GetId(),
			Verified: true,
		}
		if _, err = b.accClient.UpdateAccount(ctx, updRequest); err != nil {
			return err
		}
		if _, err = b.discoClient.NewSurveyUser(ctx, reg.GetSurveyValue()); err != nil {
			return err
		}
	}
	return nil
}

func (b *BootstrapRepo) sharedExecv2(ctx context.Context, orgs []*bsrpc.BSOrg, projects []*bsrpc.BSProject, regularAccounts []*bsrpc.BSRegularAccount) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	var err error
	err = b.resetAll(ctx)
	if err != nil {
		return err
	}

	for _, org := range orgs {
		if _, err = b.accRepo.NewOrg(ctx, org.GetOrg()); err != nil {
			return err
		}
	}
	for _, proj := range projects {
		if _, err = b.accRepo.NewProject(ctx, proj.GetProject()); err != nil {
			return err
		}
		if _, err = b.discoRepo.NewDiscoProject(ctx, proj.GetProjectDetails()); err != nil {
			return err
		}
		if proj.GetSurveySchema() != nil {
			if _, err = b.discoRepo.NewSurveyProject(ctx, proj.GetSurveySchema()); err != nil {
				return err
			}
		}
	}
	for _, reg := range regularAccounts {
		var acc *accountRpc.Account
		acc, err = b.accRepo.NewAccount(ctx, reg.GetNewAccounts())
		if err != nil {
			return err
		}
		updRequest := &accountRpc.AccountUpdateRequest{
			Id:       acc.GetId(),
			Verified: true,
		}
		if _, err = b.accRepo.UpdateAccount(ctx, updRequest); err != nil {
			return err
		}
		if _, err = b.discoRepo.NewSurveyUser(ctx, reg.GetSurveyValue()); err != nil {
			return err
		}
	}
	return nil
}

func (b *BootstrapRepo) sharedGenBS(bsAll *fakedata.BootstrapAll, joined, extension string) (string, error) {
	switch extension {
	case "json":
		jbytes, err := bsAll.MarshalPretty()
		if err != nil {
			return "", err
		}
		return joined, ioutil.WriteFile(joined, jbytes, 0644)
	case "yml", "yaml":
		ybytes, err := bsAll.MarshalYaml()
		if err != nil {
			return "", err
		}
		return joined, ioutil.WriteFile(joined, ybytes, 0644)
	default:
		return "", fmt.Errorf("invalid filename extension: %s", extension)
	}
}

func (b *BootstrapRepo) resetAll(ctx context.Context) error {
	var err error
	_, err = b.busClient.Broadcast(ctx, &coreRpc.EventRequest{
		EventName:   "onResetAllModDisco",
		Initiator:   "bootstrap-service",
		UserId:      "",
		JsonPayload: []byte{},
	})
	if err != nil {
		return err
	}
	_, err = b.busClient.Broadcast(ctx, &coreRpc.EventRequest{
		EventName:   "onResetAllSysAccount",
		Initiator:   "bootstrap-service",
		UserId:      "",
		JsonPayload: []byte{},
	})
	if err != nil {
		return err
	}
	return nil
}
