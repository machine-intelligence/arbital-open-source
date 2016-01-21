'use strict';

/* jasmine specs for controllers go here */
describe('zan controllers', function() {

	describe('ArbitalCtrl', function(){

		beforeEach(module('arbital'));

		var $controller;
		var scope, ctrl, $httpBackend;

		beforeEach(inject(function(_$controller_){
			$controller = _$controller_;
		}));

		// The injector ignores leading and trailing underscores here (i.e. _$httpBackend_).
		// This allows us to inject a service but then attach it to a variable
		// with the same name as the service in order to avoid a name conflict.
		beforeEach(inject(function(_$httpBackend_, $rootScope, $controller) {
			$httpBackend = _$httpBackend_;

			scope = $rootScope.$new();
			ctrl = $controller('ArbitalCtrl', {$scope: scope});
		}));
/*
		it('should create "karmaTestValue" model with value 3', inject(function($controller) {
			var scope = {},
			ctrl = $controller('ArbitalCtrl', {$scope:scope});

			expect(scope.karmaTestValue).toBe(3);
		}));
*/
	});
});
