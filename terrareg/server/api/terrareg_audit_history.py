
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
from terrareg.audit import AuditEvent


class ApiTerraregAuditHistory(ErrorCatchingResource):
    """Interface to obtain audit history"""

    method_decorators = [auth_wrapper('is_admin')]

    def _get(self):
        """Obtain audit history events"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'search[value]', type=str,
            required=False,
            default='',
            help='Templated URL for browsing repository.',
            dest='query'
        )
        parser.add_argument(
            'length', type=int,
            required=False,
            default=10,
            help='Module provider git tag format.'
        )
        parser.add_argument(
            'start', type=int,
            required=False,
            default=0,
            help='Path within git repository that the module exists.'
        )
        parser.add_argument(
            'order[0][dir]', type=str,
            required=False,
            default='desc',
            help='Whether module provider is marked as verified.',
            dest='order_dir'
        )
        parser.add_argument(
            'order[0][column]', type=int,
            required=False,
            help='CSRF token',
            default=0,
            dest='order_by'
        )
        parser.add_argument(
            'draw', type=int,
            required=False,
            help='draw ID',
            default=0
        )

        args = parser.parse_args()
        columns = [
            'timestamp',
            'username',
            'action',
            'object_id',
            'old_value',
            'new_value'
        ]
        order_by = 'timestamp'
        if args.order_by < len(columns):
            order_by = columns[args.order_by]

        events, total_count, filtered_count = AuditEvent.get_events(
            limit=args.length,
            offset=args.start,
            descending=args.order_dir == 'desc',
            order_by=order_by,
            query=args.query
        )

        return {
            "data": [
                {
                    'timestamp': event['timestamp'].isoformat(),
                    'username': event['username'],
                    'action': event['action'].name,
                    'object_id': event['object_id'],
                    'old_value': event['old_value'],
                    'new_value': event['new_value']
                }
                for event in events
            ],
            "draw": args.draw + 1,
            "recordsTotal": total_count,
            "recordsFiltered": filtered_count
        }
