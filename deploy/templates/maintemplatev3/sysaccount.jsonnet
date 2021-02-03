local accMixin = import "vendor/github.com/amplify-cms/sys/sys-account/service/go/template.sysaccount.libsonnet";

local cfg = {
    sysAccountConfig: accMixin.sysAccountConfig {
        unauthenticatedRoutes: accMixin.UnauthenticatedRoutes + [
            "/v2.sys_account.services.AccountService/ListAccounts",
            "/v2.sys_account.services.AccountService/GetAccount"
        ],
    }
};

std.manifestYamlDoc(cfg)