# OpenFTV Management Interface


### Build
Install dependencies
```bash
npm install
```

Run development server
```bash
npm run dev
```

### Docker images
Build the image
```bash
docker build -t management-interface-app:latest-dev .
```

Run the container on port 8080
```bash
docker run -p 8080:80 management-interface-app:latest-dev
```