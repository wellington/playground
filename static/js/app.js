var app = angular.module('app', []);

app.factory('compile', ['$http', function($http) {
  return function(d) {
    return $http.post('/compile', d);
  }
}]);
