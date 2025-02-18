# 📜 DropGox - Secure File Storage & Sync

**DropGox** is a lightweight, secure, and scalable **file storage and synchronization solution**.  
It provides a **Go-based backend**, **React frontend**, and **CLI support** for managing files.

---

## 🚀 Features
✅ **File Upload & Download** – REST API for securely storing and retrieving files.  
✅ **Authentication & Authorization** – Uses **Keycloak** for secure access control.   
✅ **CLI Support** – Command-line interface for uploading & downloading files.    
✅ **Dockerized Deployment** – Run in an isolated, containerized environment.   

---

## ⚙️ Tech Stack
- **Backend:** Go (using `mux` for routing), Keycloak for authentication  
- **Frontend:** React (WIP)
- **Database (Future Support):** PostgreSQL or an object storage system  
- **Containerization:** Docker  

---

## 📂 Project Structure
```
DropGox/
├── main.go          # Entry point for backend
├── utils/           # Utility functions (auth, file handling, etc.)
│   ├── auth.go      # Authentication functions (JWT validation)
├── Dockerfile       # Containerization setup
├── .github/workflows/ # CI/CD pipeline config
│   ├── ci-cd.yml    # GitHub Actions workflow
├── frontend/        # React-based frontend (in development)
├── cli/             # CLI for file operations (in development)
└── README.md        # You're here!
```

---

## 🛠 Setup & Installation
### **1️⃣ Clone the Repository**
```sh
git clone https://github.com/michael-oancea/DropGox.git
cd DropGox
```

### **2️⃣ Install Dependencies**
Ensure you have **Go 1.24.0+** installed:
```sh
go mod tidy
```

### **3️⃣ Run the Backend**
```sh
go run main.go
```
The API should now be running at:
```
http://localhost:8080
```

---

## 🔑 Authentication
DropGox uses **Keycloak** for authentication.  
To obtain an **access token**, send a request to the Keycloak server:

```sh
curl -X POST "http://your-auth-server.com/token"      -d "grant_type=client_credentials"      -d "client_id=dropgox-backend"      -d "client_secret=YOUR_SECRET"
```
Use this token in API requests:
```sh
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN"      http://localhost:8080/download/{filename}
```

---

## 📡 Endpoints
| Method | Endpoint                | Description          |
|--------|-------------------------|----------------------|
| `GET`  | `/health`               | Check server status |
| `POST` | `/upload`               | Upload a file       |
| `GET`  | `/download/{filename}`  | Download a file     |

Example file upload:
```sh
curl -X POST http://localhost:8080/upload      -H "Authorization: Bearer YOUR_ACCESS_TOKEN"      -F "file=@path/to/file.txt"
```

---

## 📦 Docker Support
Run the backend inside a **Docker container**:
```sh
docker build -t dropgox-backend .
docker run -p 8080:8080 dropgox-backend
```

---

## 🚀 Roadmap
- 🔹 **Frontend Development** – Implement React UI with file browsing.  
- 🔹 **Mobile App** – Android/iOS support with React Native.  
- 🔹 **Database Support** – PostgreSQL or object storage for metadata.  
- 🔹 **Encryption** – End-to-end encryption for stored files.  

---

## 🤝 Contributing
1. Fork the repository  
2. Create a new branch (`git checkout -b feature-branch`)  
3. Commit changes (`git commit -m "Add new feature"`)  
4. Push to the branch (`git push origin feature-branch`)  
5. Open a Pull Request  

---

## 🛠 Troubleshooting
### ❌ **Invalid Token Error?**
- Ensure the **access token is valid** and not expired (`exp` timestamp).  
- Verify that the token **includes the correct scopes** (`download`).  
- Check if the **API is using the correct Keycloak public key**.
---

## 📜 License
MIT License © 2024 Michael Oancea
