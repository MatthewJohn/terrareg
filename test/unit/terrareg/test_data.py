
test_data_full = {
    'testnamespace': {
        'testmodulename': {'testprovider': {
            'id': 1,
            'latest_version': '2.4.1',
            'versions': {'2.4.1': {}, '1.0.0': {}}
        }},
        'lonelymodule': {'testprovider': {
            'id': 2,
            'latest_version': '1.0.0',
            'versions': {'1.0.0': {}}
        }},
        'mock-module': {'testprovider': {
            'id': 3,
            'latest_version': '1.2.3',
            'versions': {'1.2.3': {}}
        }}
    },
    'secondtestnamespace': {
        'mockmodule2': { 'secondprovider': {
            'id': 4,
            'latest_version': '3.0.0',
            'versions': {'3.0.0': {}}
        }}
    }
}
