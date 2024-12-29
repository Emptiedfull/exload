var ctxpi = document.getElementById('serverUsagePieChart').getContext('2d');
var serverUsagePieChart = new Chart(ctxpi, {
    type: 'pie', // Set the chart type to 'pie'
    data: {
        labels: ['api', 'admin'],
        datasets: [{
            label: 'Usage',
            data: [3, 5], // Example data
            backgroundColor: [
                
                'rgba(255, 206, 86, 0.2)',
                'rgba(75, 192, 192, 0.2)'
            ],
            borderColor: [
             
                'rgba(255, 206, 86, 1)',
                'rgba(75, 192, 192, 1)'
            ],
            borderWidth: 1
        }]
    },
    options: {
        responsive: true,
        plugins: {
            legend: {
                position: 'top',
            },
            tooltip: {
                callbacks: {
                    label: function(context) {
                        let label = context.label || '';
                        if (label) {
                            label += ': ';
                        }
                        label += context.raw;
                        return label;
                    }
                }
            }
        }
    }
});
