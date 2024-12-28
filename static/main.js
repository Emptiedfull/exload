


window.addEventListener("DOMContentLoaded",()=>{

    var socket = new WebSocket("/admin/ws/main_graph")

    


    var ctx = document.getElementById('myChart').getContext('2d');
    var myChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [], // Start with an empty array for labels
            datasets: [
                {
                    label: 'Requests per second',
                    data: [], // Start with an empty array for data
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: '#9F9FF8',
                    borderWidth: 1,
                    fill: false,
                    // Add tension to make the line smooth
                    yAxisID: 'y' // Bind this dataset to the left y-axis
                },
                {
                    label: 'Memory usage (MB)',
                    data: [], // Start with an empty array for data
                    backgroundColor: 'rgba(255, 99, 132, 0.2)',
                    borderColor: 'rgba(255, 99, 132, 1)', // Different color for the second line
                    borderWidth: 1,
                    fill: false,
                   
                    yAxisID: 'y1'
                }
            ]
        },
        options: {
            scales: {
                x: {
                    type: 'time', // Use the time scale
                    time: {
                        unit: 'second', 
                        stepSize:5,
                        tooltipFormat: 'PPpp', // Format for the tooltip
                        displayFormats: {
                            second: 'HH:mm:ss' // Format for the x-axis labels
                        }
                    },
                    title: {
                        display: true,
                        text: 'Time',
                        color: '#9F9FF8' // Change the color of the x-axis title
                    },
                    ticks: {
                        source: 'auto', // Ensure ticks are generated automatically
                        autoSkip: false, // Do not skip ticks
                        maxRotation: 0, // Prevent label rotation
                        color: 'rgba(255, 255, 255, 0.40)' // Change the color of the x-axis labels
                    }
                },
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Requests',
                        color: '#9F9FF8' // Change the color of the y-axis title
                    },
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.40)' // Change the color of the y-axis labels
                    },
                    suggestedMin: 0, // Suggest a minimum value for the y-axis
                    suggestedMax: 100 // Suggest a maximum value for the y-axis
                },
                y1: {
                    beginAtZero: true,
                    position: 'right',
                    title: {
                        display: true,
                        text: 'Memory (MB)',
                        color: '#92BFFF' 
                    },
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.40)' 
                    },
                    grid: {
                        drawOnChartArea: false 
                    },
                    suggestedMin: 50, 
                    suggestedMax: 150 
                }
            },
            plugins: {
                legend: {
                    labels: {
                        color: '#FFFFFF' 
                    }
                }
            }
        }
    });

   

   

    var allData = {
        labels: [],
        rps: [],
        mem: []
    };

    // Create a WebSocket connection
    

    // Handle incoming messages
    socket.onmessage = function(event) {
        let sample = JSON.parse(event.data);

        // Store the new data in the allData array
        allData.labels.push(new Date());
        allData.rps.push(sample.rps);
        allData.mem.push(sample.mem);

        
        
        updateChart();
    };

   
    function updateChart() {
       
        console.log(allData);

        console.log( allData.labels.slice(-8))

        myChart.data.labels = allData.labels.slice(-10);
        myChart.data.datasets[0].data = allData.rps.slice(-10);
        myChart.data.datasets[1].data = allData.mem.slice(-10);

        myChart.update();
    }

})