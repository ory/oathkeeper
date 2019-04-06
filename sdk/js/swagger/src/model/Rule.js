/**
 * ORY Oathkeeper
 * ORY Oathkeeper is a reverse proxy that checks the HTTP Authorization for validity against a set of rules. This service uses Hydra to validate access tokens and policies.
 *
 * OpenAPI spec version: Latest
 * Contact: hi@ory.am
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 *
 * Swagger Codegen version: 2.2.3
 *
 * Do not edit the class manually.
 *
 */

(function(root, factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Register as an anonymous module.
    define(['ApiClient', 'model/RuleHandler', 'model/RuleMatch', 'model/Upstream'], factory);
  } else if (typeof module === 'object' && module.exports) {
    // CommonJS-like environments that support module.exports, like Node.
    module.exports = factory(require('../ApiClient'), require('./RuleHandler'), require('./RuleMatch'), require('./Upstream'));
  } else {
    // Browser globals (root is window)
    if (!root.OryOathkeeper) {
      root.OryOathkeeper = {};
    }
    root.OryOathkeeper.Rule = factory(root.OryOathkeeper.ApiClient, root.OryOathkeeper.RuleHandler, root.OryOathkeeper.RuleMatch, root.OryOathkeeper.Upstream);
  }
}(this, function(ApiClient, RuleHandler, RuleMatch, Upstream) {
  'use strict';




  /**
   * The Rule model module.
   * @module model/Rule
   * @version Latest
   */

  /**
   * Constructs a new <code>Rule</code>.
   * @alias module:model/Rule
   * @class
   */
  var exports = function() {
    var _this = this;








  };

  /**
   * Constructs a <code>Rule</code> from a plain JavaScript object, optionally creating a new instance.
   * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
   * @param {Object} data The plain JavaScript object bearing properties of interest.
   * @param {module:model/Rule} obj Optional instance to populate.
   * @return {module:model/Rule} The populated <code>Rule</code> instance.
   */
  exports.constructFromObject = function(data, obj) {
    if (data) {
      obj = obj || new exports();

      if (data.hasOwnProperty('authenticators')) {
        obj['authenticators'] = ApiClient.convertToType(data['authenticators'], [RuleHandler]);
      }
      if (data.hasOwnProperty('authorizer')) {
        obj['authorizer'] = RuleHandler.constructFromObject(data['authorizer']);
      }
      if (data.hasOwnProperty('credentials_issuer')) {
        obj['credentials_issuer'] = RuleHandler.constructFromObject(data['credentials_issuer']);
      }
      if (data.hasOwnProperty('description')) {
        obj['description'] = ApiClient.convertToType(data['description'], 'String');
      }
      if (data.hasOwnProperty('id')) {
        obj['id'] = ApiClient.convertToType(data['id'], 'String');
      }
      if (data.hasOwnProperty('match')) {
        obj['match'] = RuleMatch.constructFromObject(data['match']);
      }
      if (data.hasOwnProperty('upstream')) {
        obj['upstream'] = Upstream.constructFromObject(data['upstream']);
      }
    }
    return obj;
  }

  /**
   * Authenticators is a list of authentication handlers that will try and authenticate the provided credentials. Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive result will be the one used.  If you want the rule to first check a specific authenticator  before \"falling back\" to others, have that authenticator as the first item in the array.
   * @member {Array.<module:model/RuleHandler>} authenticators
   */
  exports.prototype['authenticators'] = undefined;
  /**
   * @member {module:model/RuleHandler} authorizer
   */
  exports.prototype['authorizer'] = undefined;
  /**
   * @member {module:model/RuleHandler} credentials_issuer
   */
  exports.prototype['credentials_issuer'] = undefined;
  /**
   * Description is a human readable description of this rule.
   * @member {String} description
   */
  exports.prototype['description'] = undefined;
  /**
   * ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you. You will need this ID later on to update or delete the rule.
   * @member {String} id
   */
  exports.prototype['id'] = undefined;
  /**
   * @member {module:model/RuleMatch} match
   */
  exports.prototype['match'] = undefined;
  /**
   * @member {module:model/Upstream} upstream
   */
  exports.prototype['upstream'] = undefined;



  return exports;
}));


