'use strict';

var appControllers = require('./index');

appControllers.controller('ReInviteController', ['$rootScope', '$scope', '$http', '$location', function($rootScope, $scope, $http, $location) {
  $scope.reInvite = function() {
    $http.post('/reinvite/execute', {
      userId: $scope.userId
    }, {
      cache: false
    }
    ).success(function(data) {
      if (data.error) {
        $rootScope.$broadcast('showAlert', data.messages);
        return;
      }
      $location.path('/reinvite/complete');
      $rootScope.$broadcast('showMessage', ['メールを送信しました。']);
    }).error(function(data) {
      $rootScope.$broadcast('showAlert', ['システムエラーが発生しました。']);
      return;
    });
  };
}]);
