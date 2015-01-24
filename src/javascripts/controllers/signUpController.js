'use strict';

var appControllers = require('./index');

appControllers.controller('SignUpController', ['$rootScope', '$scope', '$http', '$location', function($rootScope, $scope, $http, $location) {
  $scope.signUp = function() {
    var inviteCode = $location.absUrl().replace(/^.*signup\//, '').replace(/\?.*/, '').replace(/#.*/, '');
    $http.post('/signup/execute', {
      userId: $scope.userId,
      password: $scope.password,
      confirmPassword: $scope.confirmPassword,
      inviteCode: inviteCode
    }, {
      cache: false
    }
    ).success(function(data) {
      if (data.error) {
        $rootScope.$broadcast('showAlert', data.messages);
        return;
      }
      $location.path('/signup/complete');
      $rootScope.$broadcast('showMessage', ['登録しました。']);
    }).error(function(data) {
      $rootScope.$broadcast('showAlert', ['システムエラーが発生しました。']);
      return;
    });
  };
}]);
