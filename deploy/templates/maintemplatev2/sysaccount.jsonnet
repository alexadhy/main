local accMixin = import "vendor/github.com/amplify-cms/sys/sys-account/service/go/template.sysaccount.libsonnet";
local loadVar = import "vendor/github.com/amplify-cms/sys-share/sys-core/service/config/mixins/mixin.loadfn.libsonnet";

local cfg = {
    sysAccountConfig: accMixin.sysAccountConfig {
        superUserFilePath: "./config/supers.yml",
        unauthenticatedRoutes: accMixin.UnauthenticatedRoutes + [
            "/v2.sys_account.services.OrgProjService/ListOrg",
            "/v2.sys_account.services.OrgProjService/ListNonSubscribedOrgs",
            "/v2.sys_account.services.OrgProjService/ListProject",
            "/v2.sys_account.services.OrgProjService/GetOrg",
            "/v2.sys_account.services.OrgProjService/GetProject",
            "/v2.mod_disco.services.SurveyService/ListDiscoProject",
            "/v2.mod_disco.services.SurveyService/GetDiscoProject",
            "/v2.mod_disco.services.SurveyService/GenTempId",
            "/v2.mod_disco.services.SurveyService/ListSurveyProject",
            "/v2.mod_disco.services.SurveyService/GetSurveyProject",
            "/v2.mod_disco.services.SurveyService/NewSurveyUser",
        ],
        sysCoreConfig: {
           db: accMixin.CoreDB {
               name: "sysaccount.db",
               encryptKey: loadVar(prefixName="SYS_ACCOUNT", env="DB_ENCRYPT_KEY").val,
               deletePrevious: true,
           },
           cron: accMixin.CoreCron {
               backupSchedule: "@daily",
           },
        },
        sysFileConfig: {
           db: accMixin.CoreDB {
               name: "sysfiles.db",
               encryptKey: loadVar(prefixName="SYS_ACCOUNT", env="FILEDB_ENCRYPT_KEY").val,
               deletePrevious: true,
           },
           cron: accMixin.CoreCron {
               backupDir: "db/sys/sysfiles-backup",
               backupSchedule: "@daily",
           },
        },
        mailConfig: accMixin.CoreMail //{
//            sendgrid: accMixin.CoreMail.sendgrid {
//                apiKey: loadVar (prefixName="SYS_ACCOUNT", env="SENDGRID_API_KEY").val,
//            }
//        },
    }
};

std.manifestYamlDoc(cfg)
