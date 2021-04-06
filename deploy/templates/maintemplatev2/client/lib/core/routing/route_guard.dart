import 'package:flutter_modular/flutter_modular.dart';
import 'package:sys_share_sys_account_service/view/widgets/view_model/auth_nav_view_model.dart';

class NavRailGuard implements RouteGuard {
  @override
  Future<bool> canActivate(String path, ModularRoute router) async {
    var authNavVm = Modular.get<AuthNavViewModel>();
    await Future.delayed(Duration.zero, () {
      final navIdx = authNavVm.getDynamicNavIndex(path);
      if (navIdx > 0) {
        authNavVm.setCurrentNavIndex(navIdx);
      }
    });
    return true;
  }
}
