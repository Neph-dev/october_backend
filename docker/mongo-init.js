// MongoDB initialization script
print('Starting database initialization...');

// Switch to october database
db = db.getSiblingDB('october');

// Create indexes for better performance
db.companies.createIndex({ "name": 1 }, { unique: true });
db.companies.createIndex({ "ticker": 1 }, { unique: true });
db.companies.createIndex({ "industry": 1 });

db.news.createIndex({ "title": 1 });
db.news.createIndex({ "published_date": -1 });
db.news.createIndex({ "company_name": 1 });
db.news.createIndex({ "guid": 1 }, { unique: true });
db.news.createIndex({ "company_name": 1, "published_date": -1 });

// Create a simple admin user (optional)
db.createUser({
  user: "october_admin",
  pwd: "october_password_123",
  roles: [
    {
      role: "readWrite",
      db: "october"
    }
  ]
});

print('Database initialization completed.');