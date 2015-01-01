var app = angular.module('app', ['ngRoute', 'signUpControllers']);

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