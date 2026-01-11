

let cy = null;
const router = new Navigo("/");

function updateGraphData(data) {
    // Obtain graph data from relative URL
    jQuery.get(
        `/v1/terrareg/modules/${data.namespace}/${data.module}/${data.provider}/${data.version}/graph/data`,
        {
            'full_module_names': $('#graphOptionsShowFullModuleNames').is(':checked'),
            'full_resource_names': $('#graphOptionsShowFullResourceNames').is(':checked')
        }
    ).then((graphData) => {
        if (graphData) {
            cy.json({elements: graphData});
            let layout = cy.makeLayout({
                name: 'fcose',
                quality: "proof",
                avoidOverlap: true,
                fit: false,
                nodeDimensionsIncludeLabels: true,
                packComponents: false,
                randomize: false,
                animate: false,

                // Calculate node repulsion based on size of module (number of child resources)
                nodeRepulsion: node => {return node.data().child_count * 10000},

                // Disable tiling
                tile: false,

                condense:false,
            });
            layout.run();
        } else {
            $('#noGraphError').text("There is not graph data available for this module version.")
        }
    });
}

function renderPage() {
    cy = cytoscape({
        container: document.getElementById('cy'),
        style: [
            {
                selector: 'node',
                style: {
                    'label': 'data(label)',
                    'shape': 'rectangle',

                    'text-valign': 'center',
                    'text-halign': 'center',

                    'width': 'label',
                    'height': 'label',
                    'padding': '5px',

                    'border-color': '#000000',
                    'background-color': '#7681B3',
                    'color': '#FFFFFF'
                }
            }
        ],

        // Decrease wheel sensitivity to allow more granular scrolling
        wheelSensitivity: 0.5,
        zoom: 0.6,
        pan: { x: ($('#cy').width() / 2), y: ($('#cy').height() / 2) }
    })
    // Base module provider route
    router.on({
        ["/modules/:namespace/:module/:provider/:version/graph"]: {
            as: "graph",
            uses: function ({ data }) {
                updateGraphData(data);
            }
        }
    });

    router.resolve();
}
