'use strict';

var appControllers = require('./index');

appControllers.controller('MessageController', ['$scope', function($scope) {
  $scope.$on('showAlert', function(event,　messages) {
    $scope.closeMessage();
    $scope.alerts = messages;
  });

  $scope.$on('showMessage', function(event,　messages) {
    $scope.closeAlert();
    $scope.messages = messages;
  });

  $scope.closeAlert = function() {
    $scope.alerts = null;
  };

  $scope.closeMessage = function() {
    $scope.messages = null;
  };
}]);
