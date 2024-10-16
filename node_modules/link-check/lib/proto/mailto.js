"use strict";

const Isemail = require('isemail');
const LinkCheckResult = require('../LinkCheckResult');

module.exports = {
    check: function (link, opts, callback) {
        const address = link
                            .substr(7)      // strip "mailto:"
                            .split('?')[0]; // trim ?subject=blah hfields

        /* per RFC6068, the '?' is a reserved delimiter and email addresses containing '?' must be encoded,
         * so it's safe to split on '?' and pick [0].
         */

        callback(null, new LinkCheckResult(opts, link, Isemail.validate(address) ? 200 : 400, null));
    }
};
