'use strict';

const expect = require('expect.js');
const http = require('http');
const express = require('express');
const linkCheck = require('../');

describe('link-check', function () {

    this.timeout(2500);//increase timeout to enable 429 retry tests

    let baseUrl;
    let laterCustomRetryCounter;

    before(function (done) {
        const app = express();

        app.head('/nohead', function (req, res) {
            res.sendStatus(405); // method not allowed
        });
        app.get('/nohead', function (req, res) {
            res.sendStatus(200);
        });

        app.get('/foo/redirect', function (req, res) {
            res.redirect('/foo/bar');
        });
        app.get('/foo/bar', function (req, res) {
            res.json({foo:'bar'});
        });

        app.get('/loop', function (req, res) {
            res.redirect('/loop');
        });

        app.get('/hang', function (req, res) {
            // no reply
        });

        app.get('/notfound', function (req, res) {
            res.sendStatus(404);
        });

        app.get('/basic-auth', function (req, res) {

            if (req.headers.authorization === "Basic Zm9vOmJhcg==") {
                return res.sendStatus(200);
            }
            res.sendStatus(401);
        });

        // prevent first header try to be a hit
        app.head('/later-custom-retry-count', function (req, res) {
            res.sendStatus(405); // method not allowed
        });
        app.get('/later-custom-retry-count', function (req, res) {
            laterCustomRetryCounter++;

            if(laterCustomRetryCounter === parseInt(req.query.successNumber)) {
                res.sendStatus(200);
            }else{
              res.setHeader('retry-after', 1);
              res.sendStatus(429);
            }
        });

        // prevent first header try to be a hit
        app.head('/later-standard-header', function (req, res) {
            res.sendStatus(405); // method not allowed
        });
        var stdRetried = false;
        var stdFirstTry = 0;
        app.get('/later', function (req, res) {
            var isRetryDelayExpired = stdFirstTry + 1000 < Date.now();
            if(!stdRetried || !isRetryDelayExpired){
              stdFirstTry = Date.now();
              stdRetried = true;
              res.setHeader('retry-after', 1);
              res.sendStatus(429);
            }else{
              res.sendStatus(200);
            }
        });

        // prevent first header try to be a hit
        app.head('/later-no-header', function (req, res) {
            res.sendStatus(405); // method not allowed
        });
        var stdNoHeadRetried = false;
        var stdNoHeadFirstTry = 0;
        app.get('/later-no-header', function (req, res) {
            var minTime = stdNoHeadFirstTry + 1000;
            var maxTime = minTime + 100;
            var now = Date.now();
            var isRetryDelayExpired = minTime < now && now < maxTime;
            if(!stdNoHeadRetried || !isRetryDelayExpired){
              stdNoHeadFirstTry = Date.now();
              stdNoHeadRetried = true;
              res.sendStatus(429);
            }else{
              res.sendStatus(200);
            }
        });

        // prevent first header try to be a hit
        app.head('/later-non-standard-header', function (req, res) {
            res.sendStatus(405); // method not allowed
        });
        var nonStdRetried = false;
        var nonStdFirstTry = 0;
        app.get('/later-non-standard-header', function (req, res) {
            var isRetryDelayExpired = nonStdFirstTry + 1000 < Date.now();
            if(!nonStdRetried || !isRetryDelayExpired){
              nonStdFirstTry = Date.now();
              nonStdRetried = true;
              res.setHeader('retry-after', '1s');
              res.sendStatus(429);
            }else {
              res.sendStatus(200);
            }
        });

        app.get(encodeURI('/url_with_unicode–'), function (req, res) {
            res.sendStatus(200);
        });

        app.get('/url_with_special_chars\\(\\)\\+', function (req, res) {
            res.sendStatus(200);
        });

        const server = http.createServer(app);
        server.listen(0 /* random open port */, 'localhost', function serverListen(err) {
            if (err) {
                done(err);
                return;
            }
            // github action uses IPv6 addresses
            // there seems missing IPv6 support in upstream libs
            if (server.address().address === "::1") {
                baseUrl = 'http://localhost:' + server.address().port;
            } else {
                baseUrl = 'http://' + server.address().address + ':' + server.address().port;
            }
            done();
        });
    });

    it('should find that a valid link is alive', function (done) {
        linkCheck(baseUrl + '/foo/bar', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/foo/bar');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should find that a valid external link with basic authentication is alive', function (done) {
        linkCheck(baseUrl + '/basic-auth', {
            headers: {
                'Authorization': 'Basic Zm9vOmJhcg=='
            },
        }, function (err, result) {
            expect(err).to.be(null);
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should find that a valid relative link is alive', function (done) {
        linkCheck('/foo/bar', { baseUrl: baseUrl }, function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('/foo/bar');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should find that an invalid link is dead', function (done) {
        linkCheck(baseUrl + '/foo/dead', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/foo/dead');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(404);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should find that an invalid relative link is dead', function (done) {
        linkCheck('/foo/dead', { baseUrl: baseUrl }, function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('/foo/dead');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(404);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should report no DNS entry as a dead link (http)', function (done) {
        linkCheck('http://example.example.example.com/', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('http://example.example.example.com/');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(0);
            expect(result.err.code).to.be('ENOTFOUND');
            done();
        });
    });

    it('should report no DNS entry as a dead link (https)', function (done) {
        const badLink = 'https://githuuuub.com/tcort/link-check';
        linkCheck(badLink, function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(badLink);
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(0);
            expect(result.err.code).to.contain('ENOTFOUND');
            done();
        });
    });

    it('should timeout if there is no response', function (done) {
        linkCheck(baseUrl + '/hang', { timeout: '100ms' }, function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/hang');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(0);
            expect(result.err.code).to.be('ECONNRESET');
            done();
        });
    });

    it('should try GET if HEAD fails', function (done) {
        linkCheck(baseUrl + '/nohead', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/nohead');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should handle redirects', function (done) {
        linkCheck(baseUrl + '/foo/redirect', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/foo/redirect');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should handle valid mailto', function (done) {
        linkCheck('mailto:linuxgeek@gmail.com', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('mailto:linuxgeek@gmail.com');
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle valid mailto with encoded characters in address', function (done) {
        linkCheck('mailto:foo%20bar@example.org', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('mailto:foo%20bar@example.org');
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle valid mailto containing hfields', function (done) {
        linkCheck('mailto:linuxgeek@gmail.com?subject=caf%C3%A9', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('mailto:linuxgeek@gmail.com?subject=caf%C3%A9');
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle invalid mailto', function (done) {
        linkCheck('mailto:foo@@bar@@baz', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be('mailto:foo@@bar@@baz');
            expect(result.status).to.be('dead');
            done();
        });
    });

    it('should handle file protocol', function(done) {
        linkCheck('fixtures/file.md', { baseUrl: 'file://' + __dirname }, function(err, result) {
            expect(err).to.be(null);

            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle file protocol with fragment', function(done) {
        linkCheck('fixtures/file.md#section-1', { baseUrl: 'file://' + __dirname }, function(err, result) {
            expect(err).to.be(null);

            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle file protocol with query', function(done) {
        linkCheck('fixtures/file.md?foo=bar', { baseUrl: 'file://' + __dirname }, function(err, result) {
            expect(err).to.be(null);

            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle file path containing spaces', function(done) {
        linkCheck('fixtures/s p a c e/A.md', { baseUrl: 'file://' + __dirname }, function(err, result) {
            expect(err).to.be(null);

            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle baseUrl containing spaces', function(done) {
        linkCheck('A.md', { baseUrl: 'file://' + __dirname + '/fixtures/s p a c e'}, function(err, result) {
            expect(err).to.be(null);

            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            done();
        });
    });

    it('should handle file protocol and invalid files', function(done) {
        linkCheck('fixtures/missing.md', { baseUrl: 'file://' + __dirname }, function(err, result) {
            expect(err).to.be(null);

            expect(result.err.code).to.be('ENOENT');
            expect(result.status).to.be('dead');
            done();
        });
    });

    it('should ignore file protocol on absolute links', function(done) {
        linkCheck(baseUrl + '/foo/bar', { baseUrl: 'file://' }, function(err, result) {
            expect(err).to.be(null);

            expect(result.link).to.be(baseUrl + '/foo/bar');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            expect(result.err).to.be(null);
            done();
        });
    });

    it('should ignore file protocol on fragment links', function(done) {
        linkCheck('#foobar', { baseUrl: 'file://' }, function(err, result) {
            expect(err).to.be(null);

            expect(result.link).to.be('#foobar');
            done();
        });
    });

    it('should callback with an error on unsupported protocol', function (done) {
        linkCheck('gopher://gopher/0/v2/vstat', function (err, result) {
            expect(result).to.be(null);
            expect(err).to.be.an(Error);
            done();
        });
    });

    it('should handle redirect loops', function (done) {
        linkCheck(baseUrl + '/loop', function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/loop');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(0);
            expect(result.err.message).to.contain('Max redirects reached');
            done();
        });
    });

    it('should honour response codes in opts.aliveStatusCodes[]', function (done) {
        linkCheck(baseUrl + '/notfound', { aliveStatusCodes: [ 404, 200 ] },  function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/notfound');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(404);
            done();
        });
    });

    it('should honour regexps in opts.aliveStatusCodes[]', function (done) {
        linkCheck(baseUrl + '/notfound', { aliveStatusCodes: [ 200, /^[45][0-9]{2}$/ ] },  function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/notfound');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(404);
            done();
        });
    });

    it('should honour opts.aliveStatusCodes[]', function (done) {
        linkCheck(baseUrl + '/notfound', { aliveStatusCodes: [ 200 ] },  function (err, result) {
            expect(err).to.be(null);
            expect(result.link).to.be(baseUrl + '/notfound');
            expect(result.status).to.be('dead');
            expect(result.statusCode).to.be(404);
            done();
        });
    });

    it('should retry after the provided delay on HTTP 429 with standard header', function (done) {
        linkCheck(baseUrl + '/later', { retryOn429: true },  function (err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.link).to.be(baseUrl + '/later');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    it('should retry after the provided delay on HTTP 429 with non standard header, and return a warning', function (done) {
        linkCheck(baseUrl + '/later-non-standard-header', { retryOn429: true },  function (err, result) {
            expect(err).to.be(null);
            expect(result.err).not.to.be(null);
            expect(result.err).to.contain("Server returned a non standard \'retry-after\' header.");
            expect(result.link).to.be(baseUrl + '/later-non-standard-header');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    it('should retry after 1s delay on HTTP 429 without header', function (done) {
        linkCheck(baseUrl + '/later-no-header', { retryOn429: true, fallbackRetryDelay: '1s' },  function (err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.link).to.be(baseUrl + '/later-no-header');
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    // 2 is default retry so test with custom 3
    it('should retry 3 times for 429 status codes', function(done) {
        laterCustomRetryCounter = 0;
        linkCheck(baseUrl + '/later-custom-retry-count?successNumber=3', { retryOn429: true, retryCount: 3 }, function(err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    // See issue #23
    it('should handle non URL encoded unicode chars in URLs', function(done) {
        //last char is 	EN DASH
        linkCheck(baseUrl + '/url_with_unicode–', function(err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    // See issues #34 and #40
    it('should not URL encode already encoded characters', function(done) {
        linkCheck(baseUrl + '/url_with_special_chars%28%29%2B', function(err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            done();
        });
    });

    it('should support hash links', function (done) {
        linkCheck('#foo', { anchors: ['#foo'] }, function (err, result) {
            expect(err).to.be(null);
            expect(result.err).to.be(null);
            expect(result.status).to.be('alive');
            expect(result.statusCode).to.be(200);
            linkCheck('#bar', { anchors: ['#foo'] }, function (err, result) {
                expect(err).to.be(null);
                expect(result.err).to.be(null);
                expect(result.status).to.be('dead');
                expect(result.statusCode).to.be(404);
                done();
            });
        });
    });

});
