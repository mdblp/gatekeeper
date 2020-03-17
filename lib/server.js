/*
 == BSD2 LICENSE ==
 Copyright (c) 2014, Tidepool Project

 This program is free software; you can redistribute it and/or modify it under
 the terms of the associated License, which is identical to the BSD 2-Clause
 License as published by the Open Source Initiative at opensource.org.

 This program is distributed in the hope that it will be useful, but WITHOUT
 ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 FOR A PARTICULAR PURPOSE. See the License for more details.

 You should have received a copy of the License along with this program; if
 not, you can obtain one from Tidepool Project at tidepool.org.
 == BSD2 LICENSE ==
 */

'use strict';

var _ = require('lodash');
var except = require('amoeba').except;
var restify = require('restify');

var log = require('./log.js')('lib/server.js');

function resultsCb(request, response, next) {
  return function (err, results) {
    if (err != null) {
      log.info(err, 'Error on url[%s]', request.url);
      response.send(500);
    } else if (results == null) {
      response.send(404);
    } else {
      response.send(200, results);
    }

    next();
  };
}

module.exports = function (userApiClient, dataBroker, config) {
  function createServer(serverConfig) {
    log.info('Creating server[%s]', serverConfig.name);
    var app = restify.createServer(serverConfig);
    app.use(restify.plugins.queryParser());
    app.use(restify.plugins.bodyParser());
    app.use(restify.plugins.gzipResponse());

    var userApiMiddleware = require('user-api-client').middleware;
    var checkToken = userApiMiddleware.checkToken(userApiClient);
    var permissions = require('amoeba').permissions(dataBroker);

    var requireReadPermissions = function(req, res, next) {
      permissions.requireUser(req, res, next, function(req, res, next) {
        if (permissions.hasUserPermissions(req._tokendata.userid, req.params.granteeid)) {
          return next();
        }
        permissions.hasCheckedPermissions(req._tokendata.userid, req.params.userid, function(result) {
          return result.admin != null || result.custodian != null;
        }, function(error, success) {
          permissions.handleResponse(error, success, req, res, next);
        });
      });
    };

    var requireCustodian = function(req, res, next) {
      permissions.requireCustodian(req, res, next);
    };

    var normalizePermissionsBody = function(req, res, next) {
      if (Buffer.isBuffer(req.body) && req.body.length === 0) {
        req.body = {};
      }
      next();
    };

    var requireWritePermissions = function(req, res, next) {
      permissions.requireUser(req, res, next, function(req, res, next) {
        permissions.hasCheckedPermissions(req._tokendata.userid, req.params.userid, function(result) {
          return result.admin != null || result.custodian != null || (!_.isEmpty(result) && _.isEmpty(req.body));
        }, function(error, success) {
          permissions.handleResponse(error, success, req, res, next);
        });
      });
    };

/**
 * @swagger
 * components:
 *   securitySchemes:
 *     tidepoolAuth:
 *       type: apiKey
 *       in: header
 *       name: x-tidepool-session-token
 *   responses:
 *     InternalError:
 *       description: An internal problem occured
 *       content:
 *         text/plain:
 *           schema:
 *             type: string
 *     PageNotFound:
 *       description: The requested document is not found
 *       content:
 *         text/plain:
 *           schema:
 *             type: string
 *     Unauthorized:
 *       description: The requester is not authorized to perform this request
 *       content:
 *         text/plain:
 *           schema:
 *             type: string
 *   schemas:
 *     ServiceStatus:
 *       type: object
 *       properties:
 *         status: 
 *           type: string
 *         version:
 *           type: string
 *       example:
 *         status: OK
 *         version: 1.4.6
 *     List:
 *       type: object
 *       additionalProperties:
 *         schema:
 *           oneOf:
 *             - $ref: '#/components/schemas/Root'
 *             - $ref: '#/components/schemas/Permission'
 *       example:
 *         subject:
 *           root: {}
 *         userX:
 *           note: {}
 *           preview: {}
 *         userY:
 *           note: {}
 *           preview: {}
 *     Root:
 *       type: object
 *       properties:
 *         root:
 *           type: object
 *     Permission:
 *       type: object
 *       properties:
 *         note: 
 *           $ref: '#/components/schemas/Note'
 *         view:
 *           $ref: '#/components/schemas/View'
 *     Note:
 *       type: object
 *       description: a Note object, always empty
 *       example:
 *     View:
 *       type: object
 *       description: a View object, always empty
 *       example:
 */

/**
 * There is no OAS definition for this route as it would not make sense in the Styx global context
 * Most Gateekeper routes are designed to be reached through /access prefixed routes. And Styx is 
 * configured as such i.e. always forwarding requests with the /access prefix. Hence, this route can 
 * never be reached in the global context of YourLoops
 * Ideally, this route should not exist anymore and might be removed in the future
 */
    app.get('/status', function(req, res, next){
      res.send(200, {'status': 'OK', 'version': config.version});
      next();
    });

/**
 * @swagger
 * /access/status:
 *  get:
 *    summary: Request Service Status with software version
 *    description: |
 *      This route simply returns 200 with software version. 
 *      It is similar to /status
 *    responses:
 *      200:
 *        description: Service Status with software version
 *        content:
 *          application/json:
 *            schema:
 *              $ref: '#/components/schemas/ServiceStatus'
 */
    app.get('/access/status', function(req, res, next){
      res.send(200, {'status': 'OK', 'version': config.version});
      next();
    });

/**
 * @swagger
 * /access/groups/{userid}:
 *  get:
 *    summary: List of users sharing data with one subject
 *    security:
 *      - tidepoolAuth: []
 *    parameters:
 *       - in: path
 *         name: userid
 *         description: ID of the user subject
 *         required: true
 *         schema:
 *           type: string
 *    responses:
 *      200:
 *        description: List of users sharing data with the subject, including the subject himself as "root"
 *        content:
 *          application/json:
 *            schema:
 *              $ref: '#/components/schemas/List'
 *      500:
 *        description: Internal Error happened
 *        schema:
 *          $ref: '#/components/responses/InternalError'
 *      401:
 *        description: Unauthorized
 *        schema:
 *          $ref: '#/components/responses/PageNotFound'
 */
    app.get('/access/groups/:userid', checkToken, requireCustodian, function(req, res, next) {
      dataBroker.groupsForUser(req.params.userid, resultsCb(req, res, next));
    });

/**
 * @swagger
 * /access/{userid}:
 *  get:
 *    summary: List of users one subject is sharing data with
 *    security:
 *      - tidepoolAuth: []
 *    parameters:
 *       - in: path
 *         name: userid
 *         description: ID of the user subject
 *         required: true
 *         schema:
 *           type: string
 *    responses:
 *      200:
 *        description: List of users the subject is sharing data with, including the subject himself as "root"
 *        content:
 *          application/json:
 *            schema:
 *              $ref: '#/components/schemas/List'
 *      500:
 *        description: Internal Error happened
 *        schema:
 *          $ref: '#/components/responses/InternalError'
 *      401:
 *        description: Unauthorized
 *        schema:
 *          $ref: '#/components/responses/PageNotFound'
 */
    app.get('/access/:userid', checkToken, requireCustodian, function(req, res, next) {
      dataBroker.usersInGroup(req.params.userid, resultsCb(req, res, next));
    });

/**
 * @swagger
 * /access/{userid}/{granteeid}:
 *  get:
 *    summary: Check whether one subject is sharing data with one other user
 *    security:
 *      - tidepoolAuth: []
 *    parameters:
 *       - in: path
 *         name: userid
 *         required: true
 *         description: ID of the user subject
 *         schema:
 *           type: string
  *       - in: path
 *         name: granteeid
 *         required: true
 *         description: ID of the user to check for having permissions to view subject's data
 *         schema:
 *           type: string
 *    responses:
 *      200:
 *        description: Permission
 *        content:
 *          application/json:
 *            schema:
 *              $ref: '#/components/schemas/Permission'
 *      500:
 *        description: Internal Error happened
 *        schema:
 *          $ref: '#/components/responses/InternalError'
 *      404:
 *        description: No matching is found
 *        schema:
 *          $ref: '#/components/responses/PageNotFound'
 */
    app.get('/access/:userid/:granteeid', checkToken, requireReadPermissions, function(req, res, next) {
      dataBroker.userInGroup(req.params.granteeid, req.params.userid, resultsCb(req, res, next));
    });

/**
 * @swagger
 * /access/{userid}/{granteeid}:
 *  post:
 *    summary: Assign permission to one user to view subject's data
 *    description: |
 *      Create a new permission for the user to be able to see subject's data
 *      If a permission already exists, it will be updated according to request.body
 *      If the permission given in request body is empty, existing permission will be deleted
 *        i.e. user will not be able to see subject's data anymore
 *    security:
 *      - tidepoolAuth: []
 *    requestBody:
 *      description: Permission details to be applied
 *      required: true
 *      content:
 *        application/json:
 *          schema:
 *            $ref: '#/components/schemas/Permission'
 *    parameters:
 *       - in: path
 *         name: userid
 *         required: true
 *         description: ID of the user subject
 *         schema:
 *           type: string
 *       - in: path
 *         name: granteeid
 *         required: true
 *         description: ID of the user that will receive permission to view subject's data
 *         schema:
 *           type: string
 *    responses:
 *      200:
 *        description: Permission
 *        content:
 *          application/json:
 *            schema:
 *              $ref: '#/components/schemas/Permission'
 *      500:
 *        description: Internal Error happened
 *        schema:
 *          $ref: '#/components/responses/InternalError'
 *      404:
 *        description: No matching is found
 *        schema:
 *          $ref: '#/components/responses/PageNotFound'
 */
    app.post('/access/:userid/:granteeid', checkToken, normalizePermissionsBody, requireWritePermissions, function(req, res, next) {
      dataBroker.setPermissions(req.params.granteeid, req.params.userid, req.body, function(error) {
        if (!permissions.errorResponse(error, res, next)) {
          dataBroker.userInGroup(req.params.granteeid, req.params.userid, function(error, result) {
            if (!permissions.errorResponse(error, res, next)) {
              permissions.successResponse(200, result, res, next);
            }
          });
        }
      });
    });

    app.on('uncaughtException', function(req, res, route, err) {
      log.error(err, 'Uncaught exception on route[%s]!', route.spec == null ? 'unknown' : route.spec.path);
      res.send(500);
    });

    return app;
  }

  var objectsToManage = [];
  return {
    withHttp: function(port, cb){
      var server = createServer({ name: 'GatekeeperHttp' });
      objectsToManage.push(
        {
          start: function(){
            server.listen(port, function(err){
              if (err == null) {
                log.info('Http server listening on port[%s]', port);
              }
              if (cb != null) {
                cb(err);
              }
            });
          },
          close: server.close.bind(server)
        }
      );
      return this;
    },
    withHttps: function(port, config, cb){
      var server = createServer(_.extend({ name: 'GatekeeperHttps' }, config));
      objectsToManage.push(
        {
          start: function(){
            server.listen(port, function(err){
              if (err == null) {
                log.info('Https server listening on port[%s]', port);
              }
              if (cb != null) {
                cb(err);
              }
            });
          },
          close: server.close.bind(server)
        }
      );
      return this;
    },
    start: function() {
      if (objectsToManage.length < 1) {
        throw except.ISE('Gatekeeper must listen on a port to be useful, specify http, https or both.');
      }

      objectsToManage.forEach(function(obj){ obj.start(); });
      return this;
    },
    close: function() {
      objectsToManage.forEach(function(obj){ obj.close(); });
      return this;
    }
  };
};
