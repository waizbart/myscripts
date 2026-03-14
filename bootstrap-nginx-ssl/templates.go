package main

const dockerComposeTpl = `services:
{{- range .Services}}
  {{.Name}}:
    build: {{$.ProjectDir}}/{{.Name}}
    container_name: {{.Name}}
    restart: unless-stopped
    env_file:
      - {{$.ProjectDir}}/{{.Name}}/.env
    ports:
      - "{{.Port}}:{{.Port}}"
    networks:
      - app-network
{{- if $.Database.Enabled}}
    depends_on:
      - mysql
{{- end}}
{{- end}}
{{- if .Database.Enabled}}

  mysql:
    image: mysql:8
    container_name: mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: "{{.Database.RootPassword}}"
    ports:
      - "{{.Database.MySQLPort}}:3306"
    volumes:
      - {{.ProjectDir}}/mysql-data:/var/lib/mysql
    networks:
      - app-network

  phpmyadmin:
    image: phpmyadmin/phpmyadmin
    container_name: phpmyadmin
    restart: unless-stopped
    environment:
      PMA_HOST: mysql
      PMA_PORT: "3306"
    ports:
      - "{{.Database.AdminPort}}:80"
    depends_on:
      - mysql
    networks:
      - app-network
{{- end}}

networks:
  app-network:
    driver: bridge
`

const nginxSiteTpl = `server {
    listen 80;
    server_name {{.Domain}};

    location / {
        proxy_pass http://127.0.0.1:{{.Port}};
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
`
