import 'package:bootstrapper/pkg/i18n/bootstrap_localization.dart';
import 'package:bootstrapper/pkg/view/bootstrap_detail_view.dart';
import 'package:bootstrapper/pkg/view/view_model/bootstrap_view_model.dart';
import 'package:bootstrapper/rpc/v2/main_bootstrap_models.pb.dart';
import 'package:flutter/material.dart';
import 'package:flutter_modular/flutter_modular.dart';
import 'package:provider_architecture/provider_architecture.dart';
import 'package:sys_core/pkg/widgets/admin_master_details.dart';

class BootstrapView extends StatelessWidget {
  final String id;
  final String routePlaceholder;

  const BootstrapView({Key key, this.id = '', this.routePlaceholder})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ViewModelProvider.withConsumer(
      viewModelBuilder: () => Modular.get<BootstrapViewModel>(),
      onModelReady: (BootstrapViewModel model) async {
        await model.fetchBootstraps();
      },
      builder: (context, BootstrapViewModel model, child) => Scaffold(
        body: AdminMasterDetails<BS>(
          enableSearchBar: false,
          id: id,
          items: model.bootstrapList,
          labelBuilder: (item) => item.fileId,
          routeWithIdPlaceholder: routePlaceholder,
          detailsBuilder: (context, widgetId, isFullScreen) {
            return BootstrapDetailsView(
              id: widgetId,
              isLoading: model.isLoading,
              deleteCallback: () async {
                await model.refreshBootstrapList();
              },
            );
          },
          noItemsAvailable: Center(
            child: Text(
              BootstrapLocalizations.of(context).translate('noBootstraps'),
              style: Theme.of(context).copyWith().textTheme.headline4,
              textAlign: TextAlign.center,
            ),
          ),
          noItemsSelected: Center(
            child: Container(),
          ),
          disableBackButtonOnNoItemSelected: false,
          masterAppBarTitle: Text(
              BootstrapLocalizations.of(context).translate('selectBootstrap')),
          isLoadingMoreItems: false,
          hasMoreItems: false,
        ),
      ),
    );
  }
}
