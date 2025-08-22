#!/bin/bash

# Worduel Backend Environment Setup Script
# This script helps you set up your environment configuration

set -e

echo "🎯 Worduel Backend Environment Setup"
echo "======================================"

# Check if we're in the backend directory
if [ ! -f "main.go" ]; then
    echo "❌ Error: Please run this script from the backend directory"
    exit 1
fi

# Function to setup environment
setup_env() {
    local env_type=$1
    local source_file=".env.example"
    local target_file=".env.$env_type"
    
    if [ ! -f "$source_file" ]; then
        echo "❌ Error: $source_file not found"
        exit 1
    fi
    
    if [ -f "$target_file" ]; then
        echo "⚠️  $target_file already exists"
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Skipping $target_file"
            return
        fi
    fi
    
    cp "$source_file" "$target_file"
    echo "✅ Created $target_file"
    
    # Provide specific guidance for each environment
    case $env_type in
        "development")
            echo "📝 For development, you may want to edit:"
            echo "   - ALLOWED_ORIGINS (add your frontend URL)"
            echo "   - LOG_LEVEL=debug (for more verbose logging)"
            echo "   - DEBUG_MODE=true (enable debug features)"
            ;;
        "production")
            echo "📝 For production, you MUST edit:"
            echo "   - ALLOWED_ORIGINS (set to your actual domain)"
            echo "   - SENTRY_DSN (add your Sentry project DSN)"
            echo "   - LOG_LEVEL=info (reduce log volume)"
            echo "   - VALIDATE_ORIGIN=true (enable security)"
            ;;
    esac
    
    echo "   Edit with: nano $target_file"
    echo ""
}

# Main menu
echo "Which environment would you like to set up?"
echo "1) Development environment"
echo "2) Production environment"  
echo "3) Both environments"
echo "4) Exit"
echo ""

read -p "Choose an option (1-4): " choice

case $choice in
    1)
        setup_env "development"
        echo "🎉 Development environment setup complete!"
        echo "💡 Run the server with: go run main.go"
        ;;
    2)
        setup_env "production"
        echo "🎉 Production environment setup complete!"
        echo "💡 Build for production with: go build -o worduel-backend"
        ;;
    3)
        setup_env "development"
        setup_env "production"
        echo "🎉 Both environments setup complete!"
        ;;
    4)
        echo "👋 Setup cancelled"
        exit 0
        ;;
    *)
        echo "❌ Invalid option. Please choose 1-4"
        exit 1
        ;;
esac

echo ""
echo "📚 For more configuration options, see the README.md file"
echo "🔧 Environment files created. Don't forget to customize them for your needs!"