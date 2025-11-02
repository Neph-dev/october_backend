#!/bin/bash

# Database Setup Script for Remote MongoDB Clusters
# Run this once when setting up a new MongoDB cluster (Atlas, DocumentDB, etc.)

set -e

# Configuration
DB_NAME="october"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if DATABASE_URI is set
if [ -z "$DATABASE_URI" ]; then
    echo "❌ DATABASE_URI environment variable is not set."
    echo "Please set it to your remote MongoDB connection string:"
    echo "export DATABASE_URI='mongodb+srv://username:password@cluster.mongodb.net/october'"
    exit 1
fi

print_status "Setting up database indexes and structure..."
print_status "Database: $DB_NAME"
print_status "Connection: ${DATABASE_URI:0:20}..."

# Create a temporary JavaScript file for database setup
cat > /tmp/remote-db-setup.js << 'EOF'
// Remote MongoDB cluster setup script
print('🚀 Starting remote database setup...');

// Switch to october database
db = db.getSiblingDB('october');

print('📊 Creating optimized indexes...');

// Companies collection indexes
try {
    db.companies.createIndex({ "name": 1 }, { unique: true, background: true });
    print('✅ Created unique index on companies.name');
} catch (e) {
    print('⚠️  Index on companies.name might already exist: ' + e.message);
}

try {
    db.companies.createIndex({ "ticker": 1 }, { unique: true, sparse: true, background: true });
    print('✅ Created unique sparse index on companies.ticker');
} catch (e) {
    print('⚠️  Index on companies.ticker might already exist: ' + e.message);
}

try {
    db.companies.createIndex({ "industry": 1 }, { background: true });
    print('✅ Created index on companies.industry');
} catch (e) {
    print('⚠️  Index on companies.industry might already exist: ' + e.message);
}

// News collection indexes
try {
    db.news.createIndex({ "title": 1 }, { background: true });
    print('✅ Created index on news.title');
} catch (e) {
    print('⚠️  Index on news.title might already exist: ' + e.message);
}

try {
    db.news.createIndex({ "published_date": -1 }, { background: true });
    print('✅ Created descending index on news.published_date');
} catch (e) {
    print('⚠️  Index on news.published_date might already exist: ' + e.message);
}

try {
    db.news.createIndex({ "companies": 1 }, { background: true });
    print('✅ Created index on news.companies');
} catch (e) {
    print('⚠️  Index on news.companies might already exist: ' + e.message);
}

try {
    db.news.createIndex({ "guid": 1 }, { unique: true, background: true });
    print('✅ Created unique index on news.guid');
} catch (e) {
    print('⚠️  Index on news.guid might already exist: ' + e.message);
}

try {
    db.news.createIndex({ "companies": 1, "published_date": -1 }, { background: true });
    print('✅ Created compound index on news.companies + published_date');
} catch (e) {
    print('⚠️  Compound index might already exist: ' + e.message);
}

// Performance indexes for AI/search operations
try {
    db.news.createIndex({ "relevance_score": -1 }, { background: true });
    print('✅ Created index on news.relevance_score');
} catch (e) {
    print('⚠️  Index on news.relevance_score might already exist: ' + e.message);
}

try {
    db.news.createIndex({ "feed_source": 1 }, { background: true });
    print('✅ Created index on news.feed_source');
} catch (e) {
    print('⚠️  Index on news.feed_source might already exist: ' + e.message);
}

// Text search index for content
try {
    db.news.createIndex({ 
        "title": "text", 
        "summary": "text", 
        "content": "text" 
    }, { 
        background: true,
        weights: { 
            "title": 10, 
            "summary": 5, 
            "content": 1 
        },
        name: "news_text_search"
    });
    print('✅ Created text search index on news content');
} catch (e) {
    print('⚠️  Text search index might already exist: ' + e.message);
}

// Show index status
print('\n📋 Database Collections and Indexes:');
print('Companies collection indexes:');
printjson(db.companies.getIndexes());
print('\nNews collection indexes:');
printjson(db.news.getIndexes());

// Show collection stats
print('\n📊 Collection Statistics:');
try {
    var companiesCount = db.companies.countDocuments();
    var newsCount = db.news.countDocuments();
    print('Companies count: ' + companiesCount);
    print('News articles count: ' + newsCount);
} catch (e) {
    print('Could not get collection counts: ' + e.message);
}

print('\n🎉 Remote database setup completed successfully!');
EOF

# Run the setup script
print_status "Executing database setup..."
mongosh "$DATABASE_URI" --file /tmp/remote-db-setup.js

# Clean up
rm -f /tmp/remote-db-setup.js

print_success "✅ Remote MongoDB cluster setup completed!"
print_warning "💡 Keep the local mongo-init.js for developers using docker-compose"

echo
echo "📚 Usage instructions:"
echo "1. For production: Database is now ready with all indexes"
echo "2. For local development: Use 'docker-compose --profile local-db up' to include local MongoDB"
echo "3. For remote development: Use 'docker-compose up' (without local MongoDB)"