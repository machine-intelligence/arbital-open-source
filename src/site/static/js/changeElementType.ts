// This declaration needs to live in the global namespace, so it
// can't go in util.js (where this is implemented), since util.js
// now exports identifiers and as such is now an ES6 module.
interface JQuery {
	changeElementType(newType: string): JQuery;
}
