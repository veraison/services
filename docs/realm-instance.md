# What is a Realm Instance? A Developer's Guide

> **Note:** This guide focuses specifically on Realm Instances in the context of Arm's Confidential Computing Architecture (CCA).

## The Basics

Let's break down what a Realm Instance really is in everyday terms. Think of a Realm Instance like a secured VM for your application - it's a special, isolated space where your code runs safely in a virtualized environment backed by security guarantees provided by Arm's Confidential Computing Architecture (CCA).

## Key Concepts: Building Blocks of a Realm Instance

### The Birth Certificate: Realm Initial Measurements (RIM)

Think of RIM as your Realm's birth certificate - it's the first and most fundamental piece of identity:
- It's basically a hash (digest) of the code and optionally configuration of the Realm instance when it first starts up
- Works like a fingerprint - unique to your initial code setup
- Helps others verify "Yes, this is exactly the code we expect to be running"
- Every Realm Instance needs this - it's not optional

### The Name Tag: Realm Personalization Value (RPV)

Think of RPV like a name tag for your Realm - it's optional, but super useful when you need it:
- Often includes a footprint of instance-specific data supplied to the realm
- It's like giving unique names to twins - same base code and configuration, however different identities
- Without it? Your Realm is one-of-a-kind (like a custom-built tool)
- With it? You can run multiple copies of the same code and configuration (like spinning up multiple web servers)
- Perfect for when you need to scale up identical services but keep them separate

### The Journey Tracker: Realm Extensible Measurements (REM)

REMs are like measurements recorded during a car journey or path choices made when following a map:
- Works like a car's odometer and navigation system with 4 different measurement points (rem0 through rem3)
- These are passive measurements rather than active policy enforcement - they record what happened, not control what happens
- Like recording different routes taken or fuel consumption during a journey, REMs capture execution differences that a workload may have made in response to configuration it received
- These measurements can affect the workload identity and help identify state received from users

**Real-world usage examples:**
- **Configuration Response Tracking**: REMs record how a workload behaves differently based on the configuration it receives (like taking different code paths)
- **User State Identification**: REMs can capture and identify some state or data received from users during runtime
- **Execution Path Monitoring**: Like a GPS tracking different routes taken, REMs monitor the different execution paths your code follows
- **Runtime Behavior Fingerprinting**: REMs create unique signatures based on how your code actually runs, not just how it starts

## Real-World Examples

### The Custom Workshop: One RIM per Tool (No RPV needed)

Imagine you're running a workshop where each machine does something different:
- Each tool (Realm) is unique - a lathe, a drill press, a saw
- You know each tool by its shape and purpose (RIM)
- No need for extra labels because each tool is obviously different
- Real example: A microservices architecture where each service (authentication, database, API) runs in its own Realm

### The Assembly Line: Multiple Instances of the Same Tool (Using RPV)

Now picture an assembly line with multiple workers using identical tools:
- All workers use the same type of tool (same RIM)
- Each worker's tool has a unique number (RPV)
- They can all work at the same time without confusion
- **Real example**: A workload needs to do identical job in a batch [parallel processing], so on a Server, Identical N Workloads are instantiated with the same RIM. Each instance gets a unique RPV to distinguish them, allowing:
  - Parallel processing of different data sets
  - Independent scaling of each instance
  - Separate monitoring and debugging of each worker
  - Load distribution across multiple identical services

## Under the Hood: How It All Works

### The ID Card: How We Track Each Realm
```json
{
  "scheme": "ARM_CCA",
  "type": "REFERENCE_VALUE",
  "subType": "realm.reference-value",
  "attributes": {
    "vendor": "Example Vendor",             // Who made it
    "class-id": "<UUID>",                  // The model number
    "realm-initial-measurement": "...",     // The fingerprint
    "hash-alg-id": "sha-384",              // How we took the fingerprint
    "realm-personalization-value": "...",   // The serial number (if needed)
    "rem0": "...",                         // Current state check #1
    "rem1": "..."                          // Current state check #2
  }
}
```

### Trust Levels: How We Know It's Safe

Just like airport security, we have different levels of verification:
1. "Are you who you say you are?" - Checking the Realm's identity
2. "Is your code safe to run?" - Checking what's actually running inside
   - "All Clear": Everything matches and is running properly
   - "Startup Verified": Initial state is good, but still checking runtime
   - "Unknown": We don't recognize this code

### Source of Truth: Where Reference Values Come From

Just like a birth certificate is trusted because it comes from a government department, RIM reference values need to come from trusted sources where the Realm's initial code had a registration event:
- **Software Release Events**: When code is officially released and signed
- **Build System Attestations**: Automated systems that verify and record code builds
- **Developer Registrations**: Trusted registration of code fingerprints during development
- **Certificate Authorities**: Third-party validation of code authenticity

Think of it like a chain of custody - each step in your code's journey from development to deployment needs to be documented and verifiable.

## Tips from the Trenches

### Keeping Your RIMs in Order
Think of this like maintaining tools in your workshop:
- Label different tools clearly (keep RIMs distinct for different apps)
- Track when you sharpen or replace tools (version control your RIMs)
- Keep a logbook of which tool is for what job (document your RIM mappings)

### Smart RPV Management
When running multiple instances:
- Use RPVs when you need multiple copies of the same service
- Generate strong, unique RPVs (like having a robust serial number system)
- Keep track of which instance is which (maintain an RPV registry)

### Watching Your REMs
These are more like tripwire alarms than cameras:
- They can tell you when something went wrong, but not what (for that, you need event logs)
- Know what "normal" looks like (establish baseline measurements)
- Watch for anything unusual (monitor for unexpected changes)
- Set up automatic alerts (include REM checks in your security policies)

Remember: A well-organized workshop is a productive workshop. The same goes for your Realm Instances!