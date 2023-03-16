
from flask import request


def get_request_domain():
    arg_domain = request.args.get('domain')
    if arg_domain:
        return arg_domain
    return request.host

def get_request_port():
    return request.args.get('port')

def get_request_protocol():
    return request.args.get('protocol')
