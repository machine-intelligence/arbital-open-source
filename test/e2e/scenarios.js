'use strict';

/* http://docs.angularjs.org/guide/dev_guide.e2e-testing */

describe('arbital', function() {

	describe('Main Page', function() {

		it('main page should have a search box', function() {
			browser.get('');
			var searchbox = element(by.id('input-0'));
			console.log("test");
			console.log(searchbox);
			searchbox.sendKeys('Arbital');
			//console.log(browser.getTitle());
			//console.log(browser);
			//expect(browser.getTitle()).toEqual('Arbital');
		});

		it('settings page should have settings', function() {
			browser.get('settings');
			var frequency = element(by.model('userService.user.emailFrequency'));
			console.log("test");
			console.log(frequency);
			//console.log(browser.getTitle());
			//console.log(browser);
			//expect(browser.getTitle()).toEqual('Arbital');
		});
	});
});
