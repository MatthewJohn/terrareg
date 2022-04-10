
from argparse import ArgumentParser

from terrareg.server import Server


parser = ArgumentParser('terrareg')
parser.add_argument('--ssl-cert-private-key', dest='ssl_priv_key',
                    default=None,
                    help='Path to SSL private key')
parser.add_argument('--ssl-cert-public-key', dest='ssl_pub_key',
                    default=None,
                    help='Path to SSL public key')

args = parser.parse_args()

s = Server(ssl_public_key=args.ssl_pub_key, ssl_private_key=args.ssl_priv_key)
s.run()
