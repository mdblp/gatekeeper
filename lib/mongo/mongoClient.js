/*
 * == BSD2 LICENSE ==
 * Copyright (c) 2014, Tidepool Project
 * 
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the associated License, which is identical to the BSD 2-Clause
 * License as published by the Open Source Initiative at opensource.org.
 * 
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE. See the License for more details.
 * 
 * You should have received a copy of the License along with this program; if
 * not, you can obtain one from Tidepool Project at tidepool.org.
 * == BSD2 LICENSE ==
 */

'use strict';

var _ = require('lodash');
var mongodb = require('mongodb');
var pre = require('amoeba').pre;

var getDatabaseName = require('amoeba').mongoUtil.getDatabaseName;
var log = require('../log.js')('mongo/mongoClient.js');

module.exports = function(config){
  config = _.clone(config);
  pre.hasProperty(config, 'connectionString');
  pre.defaultProperty(config, 'closeDelay', 60 * 1000);
  pre.defaultProperty(config, 'closePollDelay', 5000);

  var mongoClient = null;
  var started = false;
  let connecting = false;
  let loopConnectIntervalId;
  var requestCount = 0;
  var databaseName = 'gatekeeper';

  function closeWhenRequestsDone(clientRef, timeElapsed, cb) {
    if (requestCount > 0) {
      if (timeElapsed >= config.closeDelay) {
        log.warn('Forcibly closing mongo because operations didn\'t clear out within [%s] millis', config.closeDelay);
        clientRef.close(cb);
      } else {
        setTimeout(
          closeWhenRequestsDone.bind(null, clientRef, timeElapsed + config.closePollDelay, cb),
          config.closePollDelay
        );
      }
    } else {
      log.info('Closing mongo client');
      clientRef.close(cb);
    }
  }

  return {
    /**
     * Provides the requested collection object from Mongo to the collCb.
     *
     * The function signature on this method is a little funky because it attempts to do
     * request counting in order to handle the close() method gracefully.  The collCb will be
     * passed a callback that it *must* use.  This callback wraps the clientCb passed in here,
     * but decrements the request counter before calling it.
     *
     * Not using the callback passed to collCb will result in incorrect counts and elongated
     * `close()` times.
     *
     * @param collection String name of the mongo collection to get
     * @param clientCb Callback from the user, if an error occurs, this will short-circuit to this callback directly.
     * @param collCb callback that is passed the collection as the first argument and a callback as the second argument,
     *   the callback passed into collCb *must* be called from collCb
     * @returns {*}
     */
    withCollection: function(collectionName, clientCb, collCb) {
      if (! started) {
        return clientCb(new Error('Must start the mongoClient before using it.'));
      }

      if (mongoClient == null) {
        return clientCb(new Error('Client closing, cannot be used.'));
      }

      mongoClient.db(databaseName).collection(collectionName, function(err, collection) {
        if (err != null) {
          clientCb(err);
        }

        var requestComplete = false;
        ++requestCount;
        collCb(collection, function() {
          if (! requestComplete) {
            requestComplete = true;
            --requestCount;
          }
        });
      });
    },
    start :function(cb) {
      this.connect((err)=>{
        if(typeof cb === 'function') {
          cb(err);
        }
        if (err != null) {
          log.warn(err, 'Error on initial connection to mongo!');
          loopConnectIntervalId = setInterval(()=>this.connect(cb),35000);
          return;
        }
        log.info('Successfully connected to mongo');
      });
    },
    connect: function(cb){
      if (cb == null) {
        cb = function(err) {
          if (err != null) {
            log.warn(err, 'Error connection to mongo!');
            return;
          }
          log.info('Successfully connected to mongo');
        };
      }

      if (started) {
        return cb(new Error('Already started'));
      }
      if(connecting) {
        return cb(new Error('Connection in progress'));
      }
      connecting = true;
      requestCount = 0;

      const client = new mongodb.MongoClient(config.connectionString, { useNewUrlParser: true , useUnifiedTopology: true});
      databaseName = getDatabaseName(config.connectionString, 'gatekeeper');
      client.connect(function(err){
        connecting = false;
        if(err === null) {
          if (client != null) {
            if (mongoClient != null) {
              client.close();
              return;
            }
            started = true;
            mongoClient = client;
            if(loopConnectIntervalId !== null) {
              clearInterval(loopConnectIntervalId);
            }
          }
        }
        cb(err);
      });
    },
    close: function(cb) {
      if(loopConnectIntervalId !== null) {
        clearInterval(loopConnectIntervalId);
      }
      if (started) {
        closeWhenRequestsDone(mongoClient, 0, function(err, results){
          started = false;
          if (cb != null) {
            cb(err, results);
          }
        });

        mongoClient = null;
      }
    }
  };
};
