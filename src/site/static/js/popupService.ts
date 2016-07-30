'use strict';

import app from './angular.ts';
import {escapeHtml} from './util.ts';

// Takes care of all popup and toast related functionality
app.service('popupService', function($http, $compile, $location, $mdToast, $rootScope, $interval) {
	var that = this;

	var $popupDiv = $('#popup-div');
	var $popupHeader = $('#popup-header');
	var $popupBody = $('#popup-body');
	var popupHideCallback = undefined;
	var popupIntervalPromise = undefined;
	this.popupPercentLeft = 0;
	this.popupParams = undefined;

	// Show the popup.
	// params = {
	//	title: string to set the window title to
	//	$element: compiled element to add to the window body
	//	persistent: if true, this popup will persist when user moves between pages
	//	timeout: optional number of seconds to wait before automatically hiding the window
	// }
	this.showPopup = function(params, hideCallback) {
		if (that.popupParams) {
			that.hidePopup();
		}
		$popupBody.append(params.$element);
		$popupHeader.text(params.title);
		popupHideCallback = hideCallback;
		that.popupParams = params;
		if (params.timeout) {
			// Compute how often we need to decrease the bar by 1 percent
			var interval = params.timeout / 100;
			popupIntervalPromise = $interval(that.updatePopupTimeLeft, interval);
			that.popupPercentLeft = 100;
		}
	};

	// Called every so often to update how much time the popup has left.
	this.updatePopupTimeLeft = function() {
		that.popupPercentLeft--;
		if (that.popupPercentLeft <= 0) {
			that.hidePopup();
		}
	};

	// Hide the popup.
	this.hidePopup = function(result) {
		result = result || {};
		if (popupIntervalPromise) {
			$interval.cancel(popupIntervalPromise);
			popupIntervalPromise = undefined;
			that.popupPercentLeft = 0;
		}
		if (popupHideCallback) {
			popupHideCallback(result);
			popupHideCallback = undefined;
		}
		$popupBody.empty();
		that.popupParams = undefined;
	};

	// This is called when we go to a different page. If there is an existing popup
	// that's not persistent, hide it.
	this.hideNonpersistentPopup = function() {
		if (that.popupParams && !that.popupParams.persistent) {
			that.hidePopup();
		}
	};

	// Show an NG toast
	// params = {
	//	text: text to show
	//	scope: scope to assign to the md-toast,
	//	normalButton: {text: button text, icon: the icon to show on the button, callbackText: function to call if clicked}
	//	isError: if true, this will be an error toast
	// }
	this.showToast = function(params) {
		var toastClass = 'md-toast-content';
		if (params.isError) {
			toastClass += ' md-warn';
		}
		var hideDelay = Math.max(3000, params.text.length / 10 * 1000);
		if (params.normalButton) {
			hideDelay += 3000;
		}
		var templateHtml = '<md-toast><div class=\'' + toastClass + '\'>';
		templateHtml += '<span flex>' + escapeHtml(params.text) + '</span>';
		if (params.normalButton) {
			templateHtml += '<md-button class="md-action" ng-click="' + params.normalButton.callbackText + '">';
			templateHtml += '<span>' + escapeHtml(params.normalButton.text) + '</span>';
			if (params.normalButton.icon) {
				templateHtml += '&nbsp;<md-icon>' + escapeHtml(params.normalButton.icon) + '</md-icon>';
			}
			templateHtml += '</md-button>';
		}
		templateHtml += '</div></md-toast>';
		var toastOptions: any = {
			template: templateHtml,
			autoWrap: false,
			parent: $('#fixed-overlay'),
			preserveScope: !!params.scope,
			hideDelay: hideDelay,
		};
		if (params.scope) {
			toastOptions.scope = params.scope;
		}
		$mdToast.show(toastOptions);
	};
});
