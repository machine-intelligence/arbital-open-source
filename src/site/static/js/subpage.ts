'use strict';

import app from './angular.ts';

// Directive for showing a subpage.
app.directive('arbSubpage', function($compile, $timeout, $location, $mdToast, $mdMedia, $anchorScroll, arb, RecursionHelper) {
	return {
		templateUrl: versionUrl('static/html/subpage.html'),
		scope: {
			pageId: '@',  // id of this subpage
			lensId: '@',  // id of the lens this subpage belongs to
			parentSubpageId: '@',  // id of the parent subpage, if there is one
			showEvenIfResolved: '=',
			isTopLevelComment: '=',
			// If true, show a reply button (used when showing the subpage in the updates panel)
			showQuickReply: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.lens = arb.stateService.pageMap[$scope.lensId];
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.page.subpageIds = $scope.page.commentIds;
			$scope.page.subpageIds.sort(arb.pageService.getChildSortFunc('oldestFirst'));
			$scope.isCollapsed = false;
			$scope.showNewCommentLoading = false;
			$scope.isTinyScreen = !$mdMedia('gt-xs');

			// Check if the user has the permissions to reply to this comment. Permission
			// can come from the comment or the lens
			$scope.canReply = function() {
				if (!$scope.page.permissions || !$scope.lens.permissions) return false;
				return $scope.page.permissions.comment.has || $scope.lens.permissions.comment.has;
			};

			// TODO: This should be refactored into getPageUrl
			var url = arb.urlService.getPageUrl($scope.lensId);
			var hashIndex = url.indexOf('#');
			if (hashIndex > 0) {
				url = url.slice(0, hashIndex);
			}
			if (url.indexOf('?') < 0) {
				// We have to set the lens explicitly, so we don't get automatically
				// redirected to an easier lens or something.
				url += '?l=' + $scope.lensId;
			}
			$scope.myUrl = url + '#subpage-' + $scope.page.pageId;

			// Check if this comment is selected via URL hash
			$scope.isSelected = function() {
				return $location.hash() === 'subpage-' + $scope.page.pageId;
			};

			// Return true if this comment should be shown
			$scope.isSubpageVisible = function() {
				if ($scope.isDeleted) return false;
				if ($scope.page.isResolved && !$scope.isSelected() && !$scope.showEvenIfResolved) return false;
				return true;
			};

			// Return true iff the user can approve this comment.
			$scope.canApproveComment = function() {
				return arb.userService.userCanApproveComments($scope.lens.editDomainId);
			};

			// Approve this comment.
			$scope.approveComment = function() {
				if (!$scope.canApproveComment()) return;
				$scope.page.isApprovedComment = true;
				arb.stateService.postDataWithoutProcessing('/approveComment/', {commentID: $scope.pageId});
			};

			// Add the author of this comment to the domain.
			$scope.addUserToDomain = function() {
				$scope.approveComment();
				var data = {
					domainId: $scope.lens.editDomainId,
					userInput: $scope.page.pageCreatorId,
				};
				arb.stateService.postDataWithoutProcessing('/newMember/', data, function() {
					arb.popupService.showToast({
						text: 'User added to the domain.',
					});
				});
			};

			// Called when the user collapses/expands this subpage
			$scope.collapseToggle = function() {
				$scope.isCollapsed = !$scope.isCollapsed;
			};

			// Called when the user wants to edit the subpage
			$scope.editSubpage = function(event) {
				if (!event.ctrlKey) {
					arb.pageService.loadEdit({
						pageAlias: $scope.page.pageId,
						success: function() {
							$scope.editing = true;
						},
					});
					event.preventDefault();
				}
			};

			// Called when the user is done editing the subpage
			$scope.editDone = function(result) {
				if (!result.discard) {
					arb.pageService.newCommentCreated(result.pageId);
				}
				$scope.editing = false;
			};

			// Called when the user wants to delete the subpage
			$scope.deleteSubpage = function() {
				arb.pageService.deletePage($scope.page.pageId, function() {
					$scope.isDeleted = true;
					arb.popupService.showToast({text: 'Comment deleted'});
				}, function(data) {
					$scope.addMessage('delete', 'Error deleting page: ' + data, 'error');
				});
			};

			// Called to create a new reply
			$scope.newReply = function() {
				$scope.showNewCommentLoading = true;
				arb.pageService.newComment({
					parentPageId: $scope.lensId,
					replyToId: $scope.page.pageId,
					success: function(newCommentId) {
						$scope.newReplyId = newCommentId;
						$scope.showNewCommentLoading = false;
					},
				});
			};

			// Called when the user is done with the new reply
			$scope.newReplyDone = function(result) {
				$scope.newReplyId = undefined;
				if (!result.discard) {
					arb.pageService.newCommentCreated(result.pageId);
				}
			};

			// Called to set the comment's isEditorComment
			$scope.showToEditorsOnly = function() {
				// TODo: need to refactor the /approveComment/ end-point to accept more various changes
				$scope.page.isEditorComment = true;
				arb.pageService.savePageInfo($scope.page);
			};

			// Resolve the comment thread
			$scope.resolveThread = function() {
				arb.pageService.resolveThread($scope.pageId);
			};
		},
		compile: function(element) {
			var link = RecursionHelper.compile(element);
			var recursivePostLink = link.post;
			link.post = function(scope, element, attrs) {
				// Scroll to the subpage if it's the current hash
				if ($location.hash() == 'subpage-' + scope.pageId) {
					$timeout(function() {
						$anchorScroll();
					});
				}
				recursivePostLink(scope, element, attrs);
			};
			return link;
		}
	};
});

// Directive for container holding an inline comment
app.directive('arbInlineComment', function($compile, $timeout, $location, $mdToast, arb, RecursionHelper) {
	return {
		templateUrl: versionUrl('static/html/inlineComment.html'),
		scope: {
			commentId: '@',
			lensId: '@',  // id of the lens this comment belongs to
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isExpanded = false;
			$scope.toggleExpand = function() {
				$scope.isExpanded = !$scope.isExpanded;
			};
		},
		link: function(scope: any, element, attrs) {
			var content = element.find('.inline-subpage');
			scope.showExpandButton = function() {
				return content.get(0).scrollHeight > content.height() && !scope.isExpanded;
			};
		},
	};
});
