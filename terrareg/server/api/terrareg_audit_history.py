
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.audit


class ApiTerraregAuditHistory(ErrorCatchingResource):
    """Interface to obtain audit history"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _get(self):
        """Obtain audit history events"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'search[value]', type=str,
            required=False,
            default='',
            location='args',
            help='Templated URL for browsing repository.',
            dest='query'
        )
        parser.add_argument(
            'length', type=int,
            required=False,
            default=10,
            location='args',
            help='Module provider git tag format.'
        )
        parser.add_argument(
            'start', type=int,
            required=False,
            default=0,
            location='args',
            help='Path within git repository that the module exists.'
        )
        parser.add_argument(
            'order[0][dir]', type=str,
            required=False,
            default='desc',
            location='args',
            help='Whether module provider is marked as verified.',
            dest='order_dir'
        )
        parser.add_argument(
            'order[0][column]', type=int,
            required=False,
            help='CSRF token',
            default=0,
            location='args',
            dest='order_by'
        )
        parser.add_argument(
            'draw', type=int,
            required=False,
            help='draw ID',
            location='args',
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

        events, total_count, filtered_count = terrareg.audit.AuditEvent.get_events(
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
