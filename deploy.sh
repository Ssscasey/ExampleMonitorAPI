#!/bin/bash

sudo apt-get update
sudo apt-get upgrade -y

sudo apt-get install -y apache2


cp -r your_project_directory /var/www/html/


sudo chown -R www-data:www-data /var/www/html/your_project_directory
sudo chmod -R 755 /var/www/html/your_project_directory


sudo systemctl restart apache2

echo "Deployment completed successfully."