'use strict';

var angular = require('angular');

require('angular-route');
require('./controllers/index');

var app = angular.module('app', ['ngRoute', 'appControllers', 'ui.bootstrap']);

app.config(['$routeProvider', function($routeProvider) {
  $routeProvider.
    when('/', {
      templateUrl: '/views/signup/new.html',
      controller: 'SignUpController'
    }).
    when('/signup/complete', {
      templateUrl: '/views/signup/complete.html'
    }).
    otherwise({
      redirectTo: '/'
    });
  }
]);