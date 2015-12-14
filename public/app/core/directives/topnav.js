define([
  '../core_module',
],
function (coreModule) {
  'use strict';

  coreModule.directive('topnav', function($rootScope, contextSrv) {
    return {
      restrict: 'E',
      transclude: true,
      scope: {
        title: "@",
        section: "@",
        titleAction: "&",
        subnav: "=",
      },
      template:
        '<div class="navbar navbar-static-top"><div class="navbar-inner"><div class="container-fluid">' +
        '<div class="top-nav">' +
        '<div ng-if="!contextSrv.sidemenu" style="width:200px; float:left;">' +
        '<a class="pointer" style="display:block; padding:0; margin:8px 0 4px 22px; font-size:16px;">' +
        '<img class="logo-icon" src="img/fav32.png" ng-click="contextSrv.redirectToHome()" ' +
        'style="margin-top:5px; margin-left:5px; border-radius:50%; width:30px;"></img> ' +
        '<i class="pull-right fa fa-angle-right" ng-click="contextSrv.toggleSideMenu()" ' +
        'style="opacity:1; padding-right:5px; padding-top:5px; font-size:170%;"></i>' +
        '</a>' +
        '</div>' +

        '<span class="icon-circle top-nav-icon">' +
        '<i ng-class="icon"></i>' +
        '</span>' +

        '<span ng-show="section">' +
        '<span class="top-nav-title">{{section}}</span>' +
        '<i class="top-nav-breadcrumb-icon fa fa-angle-right"></i>' +
        '</span>' +

        '<a ng-click="titleAction()" class="top-nav-title">' +
        '{{title}}' +
        '</a>' +
        '<i ng-show="subnav" class="top-nav-breadcrumb-icon fa fa-angle-right"></i>' +
        '</div><div ng-transclude></div></div></div></div>',
      link: function(scope, elem, attrs) {
        scope.icon = attrs.icon;
        scope.contextSrv = contextSrv;

        scope.toggle = function() {
          $rootScope.appEvent('toggle-sidemenu');
        };
      }
    };
  });

});
