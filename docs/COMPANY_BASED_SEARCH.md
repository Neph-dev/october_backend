# Company-Based Web Search Implementation

## New Approach

The web search validation has been completely redesigned to be **company-based** rather than keyword-based, as requested.

## How It Works Now

### 1. **Company Database Validation**
- ✅ **If query is about a company in our database** → Allow web search
- ❌ **If query is about a company NOT in our database** → Reject web search

### 2. **No Keyword Filtering**
- ✅ **Any question** about database companies is allowed
- ✅ **Let OpenAI's intelligence** handle the appropriateness of responses
- ✅ **More permissive filtering** - trust OpenAI to provide good answers

### 3. **Company Detection**
The system detects companies through:
- **Explicit company context** passed in the request
- **Company name extraction** from the query text
- **Company variations** (RTX → Raytheon Technologies, etc.)

## Examples That Now Work

### ✅ **Any RTX Questions** (RTX is in database):
- "Who was the founder of RTX?"
- "What is RTX's favorite color?" 
- "How tall is the RTX building?"
- "What does RTX do on weekends?"
- "RTX stock price history"

### ✅ **Any Raytheon Technologies Questions** (in database):
- "What does Raytheon Technologies sell?"
- "Raytheon company culture"
- "Who is the CEO of Raytheon?"
- "Raytheon office locations"

### ✅ **Any US War Department Questions** (in database):
- "What is the mission of the US War Department?"
- "US War Department budget"
- "War Department organizational structure"

### ❌ **Non-Database Company Questions**:
- "Who founded Apple Inc?" (Apple not in database)
- "Google company history" (Google not in database)
- "Microsoft products" (Microsoft not in database)

## Technical Implementation

### 1. **Company Service Integration**
```go
type WebSearchService struct {
    client         *http.Client
    logger         *slog.Logger
    companyService company.Service  // NEW: Direct database access
    searchEngines  []SearchEngine
}
```

### 2. **Database Company Check**
```go
func (s *WebSearchService) isCompanyInDatabase(ctx context.Context, companyName string) bool {
    _, err := s.companyService.GetCompanyByName(ctx, companyName)
    return err == nil
}
```

### 3. **Company Name Extraction**
```go
companyPatterns := map[string][]string{
    "Raytheon Technologies": {"rtx", "raytheon", "raytheon technologies"},
    "US War Department": {"war department", "us war department", "war dept"},
    "Lockheed Martin": {"lockheed", "lockheed martin", "lmt"},
}
```

### 4. **Permissive Content Filtering**
```go
// More permissive - let OpenAI handle appropriateness
func (s *WebSearchService) isCompanyRelated(result WebSearchResult, companies []string) bool {
    // ... check for company mentions ...
    
    // If we can't find company mentions, accept it anyway to let OpenAI filter
    return true  // More permissive approach
}
```

## Benefits

1. **Simpler Logic**: No complex keyword filtering
2. **More Flexible**: Any question about database companies works
3. **AI-Powered**: Let OpenAI's intelligence handle response appropriateness
4. **Database-Driven**: Directly tied to companies we actually have data for
5. **User-Friendly**: Users don't need to worry about specific keywords

## Testing

Run the comprehensive test:
```bash
./scripts/test_company_based_search.sh
```

This verifies:
- ✅ RTX questions work (any question)
- ✅ Raytheon Technologies questions work
- ✅ US War Department questions work
- ❌ Non-database companies are rejected
- ✅ OpenAI provides intelligent responses regardless of question type