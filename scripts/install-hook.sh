#!/bin/bash

# Копируем hook в .git/hooks/
cp scripts/pre-commit .git/hooks/pre-commit

# Делаем его исполняемым
chmod +x .git/hooks/pre-commit

echo "✅ Pre-commit hook установлен"
echo ""
echo "Не забудьте установить переменную окружения:"
echo "  export OPENAI_API_KEY='your-api-key-here'"