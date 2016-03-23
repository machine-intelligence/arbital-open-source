'use strict';

// Directive for the User page.
app.directive('arbUserPage', function(pageService, userService) {
  return {
    templateUrl: 'static/html/userPage.html',
    scope: {
      userId: '@',
      idsMap: '=',
    },
    controller: function($scope) {
      $scope.pageService = pageService;
      $scope.userService = userService;
    },
  };
});
