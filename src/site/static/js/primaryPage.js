'use strict';

// Directive for the entire primary page.
app.directive('arbPrimaryPage', function($compile, $location, $timeout,
  pageService, userService, autocompleteService) {
  return {
    templateUrl: 'static/html/primaryPage.html',
    scope: {
      userId: '@',
      idsMap: '=',
    },
    controller: function($scope) {
      $scope.pageService = pageService;
      $scope.userService = userService;
      $scope.page = pageService.primaryPage;
      $scope.page.childIds.sort(
        pageService.getChildSortFunc($scope.page.sortChildrenBy));
      $scope.page.relatedIds.sort(pageService.getChildSortFunc('likes'));
      $scope.page.answerIds.sort(pageService.getChildSortFunc('likes'));

      // Create the edit section for a new answer
      var createNewAnswer = function() {
        $scope.newAnswerId = undefined;
        if ($scope.page.childDraftId === '') {
          pageService.getNewPage({
            type: 'answer',
            parentIds: [$scope.page.pageId],
            success: function(newAnswerId) {
              $scope.newAnswerId = newAnswerId;
            },
          });
        } else {
          pageService.loadEdit({
            pageAlias: $scope.page.childDraftId,
            success: function() {
              $scope.newAnswerId = $scope.page.childDraftId;
            },
          });
        }
      };
      if ($scope.page.type === 'question') {
        createNewAnswer();
      }

      // Called when the user is done editing the new answer
      $scope.answerDone = function(result) {
        if (result.discard) {
          createNewAnswer();
        } else {
          window.location.href = pageService.getPageUrl($scope.page.pageId) +
            '#subpage-' + $scope.newAnswerId;
          window.location.reload();
        }
      };

      // Called when the user selects an answer to suggest
      $scope.suggestedAnswer = function(result) {
        if (!result) { return; }
        var data = {
          parentId: $scope.page.pageId,
          childId: result.label,
          type: 'parent',
        };
        pageService.newPagePair(data, function() {
          window.location.href = pageService.getPageUrl($scope.page.pageId) + '#subpage-' + result.label;
          window.location.reload();
        });
      };
    },
  };
});
