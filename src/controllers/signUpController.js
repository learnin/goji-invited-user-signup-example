'use strict';

var appControllers = require('./index');

appControllers.controller('SignUpController', ['$scope', '$http', '$location', function($scope, $http, $location) {
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
      if (data.Error) {
        $scope.messages = data.Messages;
        return;
      }
      $location.path('/signup/complete');
    }).error(function(data) {
      $scope.message = "システムエラーが発生しました。";
      return;
    });
  };
}]);
