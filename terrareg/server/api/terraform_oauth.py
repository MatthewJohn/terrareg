from urllib.parse import urlencode, parse_qs

import flask
from flask import Blueprint, redirect, current_app, jsonify, session
from flask.helpers import make_response
from flask.templating import render_template
from oic.oic.message import TokenErrorResponse, UserInfoErrorResponse, EndSessionRequest

from pyop.access_token import AccessToken, BearerTokenError
from pyop.exceptions import InvalidAuthenticationRequest, InvalidAccessToken, InvalidClientAuthentication, OAuthError, \
    InvalidSubjectIdentifier, InvalidClientRegistrationRequest
from pyop.util import should_fragment_encode

from terrareg.terraform_idp import TerraformIdp
import terrareg.auth


terraform_oidc_provider_blueprint = Blueprint('terraform_oidc_provider', __name__, url_prefix='/terraform/oauth')


@terraform_oidc_provider_blueprint.route('/authorization', methods=['GET'])
def authorization_endpoints():
    # parse authentication request
    try:
        args = dict(flask.request.args)
        # terraform does not provide a 'scope', so forcefully set one
        args['scope'] = "openid"
        auth_req = TerraformIdp.get().provider.parse_authentication_request(
            urlencode(args),
            flask.request.headers
        )
    except InvalidAuthenticationRequest as e:
        current_app.logger.debug('received invalid authn request', exc_info=True)
        error_url = e.to_error_url()
        if error_url:
            return redirect(error_url, 303)
        else:
            # show error to user
            return make_response('Something went wrong: {}'.format(str(e)), 400)

    # Set authentication request session data
    session['authn_req'] = auth_req.to_dict()
    session.modified = True

    # Check if user is authenticated
    current_auth_method = terrareg.auth.AuthFactory().get_current_auth_method()
    if current_auth_method.is_authenticated():
        authn_response = TerraformIdp.get().provider.authorize(auth_req, current_auth_method.get_username())
        response_url = authn_response.request(auth_req['redirect_uri'], should_fragment_encode(auth_req))
        return redirect(response_url, 303)
    
    # Otherwise, redirect user to login page
    else:
        raise Exception("Need to redirect user")


@terraform_oidc_provider_blueprint.route('/.well-known/openid-configuration')
def provider_configuration():
    return jsonify(TerraformIdp.get().provider.provider_configuration.to_dict())


@terraform_oidc_provider_blueprint.route('/jwks')
def jwks_uri():
    return jsonify(TerraformIdp.get().provider.jwks)


@terraform_oidc_provider_blueprint.route('/token', methods=['POST'])
def token_endpoint():
    try:
        token_response = TerraformIdp.get().provider.handle_token_request(flask.request.get_data().decode('utf-8'),
                                                                   flask.request.headers)
        return jsonify(token_response.to_dict())
    except InvalidClientAuthentication as e:
        current_app.logger.debug('invalid client authentication at token endpoint', exc_info=True)
        error_resp = TokenErrorResponse(error='invalid_client', error_description=str(e))
        response = make_response(error_resp.to_json(), 401)
        response.headers['Content-Type'] = 'application/json'
        response.headers['WWW-Authenticate'] = 'Basic'
        return response
    except OAuthError as e:
        current_app.logger.debug('invalid request: %s', str(e), exc_info=True)
        error_resp = TokenErrorResponse(error=e.oauth_error, error_description=str(e))
        response = make_response(error_resp.to_json(), 400)
        response.headers['Content-Type'] = 'application/json'
        return response


@terraform_oidc_provider_blueprint.route('/userinfo', methods=['GET', 'POST'])
def userinfo_endpoint():
    try:
        response = TerraformIdp.get().provider.handle_userinfo_request(flask.request.get_data().decode('utf-8'),
                                                                flask.request.headers)
        return jsonify(response.to_dict())
    except (BearerTokenError, InvalidAccessToken) as e:
        error_resp = UserInfoErrorResponse(error='invalid_token', error_description=str(e))
        response = make_response(error_resp.to_json(), 401)
        response.headers['WWW-Authenticate'] = AccessToken.BEARER_TOKEN_TYPE
        response.headers['Content-Type'] = 'application/json'
        return response


def do_logout(end_session_request):
    try:
        TerraformIdp.get().provider.logout_user(end_session_request=end_session_request)
    except InvalidSubjectIdentifier as e:
        return make_response('Logout unsuccessful!', 400)

    redirect_url = TerraformIdp.get().provider.do_post_logout_redirect(end_session_request)
    if redirect_url:
        return redirect(redirect_url, 303)

    return make_response('Logout successful!')


@terraform_oidc_provider_blueprint.route('/logout', methods=['GET', 'POST'])
def end_session_endpoint():
    if flask.request.method == 'GET':
        # redirect from RP
        end_session_request = EndSessionRequest().deserialize(urlencode(flask.request.args))
        flask.session['end_session_request'] = end_session_request.to_dict()
        return render_template('logout.jinja2')
    else:
        form = parse_qs(flask.request.get_data().decode('utf-8'))
        if 'logout' in form:
            return do_logout(EndSessionRequest().from_dict(flask.session['end_session_request']))
        else:
            return make_response('You chose not to logout')