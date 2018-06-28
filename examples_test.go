package bloom_test

import (
	"fmt"
	"log"

	"github.com/marselester/bloom"
)

func Example() {
	const maxEmails = 100
	const prob = 0.01
	bf, err := bloom.New(maxEmails, prob)
	if err != nil {
		log.Fatalf("Bloom filter is not created: %v", err)
	}

	bob := []byte("bob@example.com")
	if err = bf.Add(bob); err != nil {
		log.Fatalf("Bloom filter couldn't add element: %v", err)
	}

	isIn, err := bf.Has(bob)
	if err != nil {
		log.Fatalf("Bloom filter couldn't test element: %v", err)
	}
	if isIn {
		fmt.Println("Bob's email is likely in the set.")
	} else {
		fmt.Println("Bob's email is not found.")
	}
	// Output:
	// Bob's email is likely in the set.
}

// Optimistic usage of a Bloom filter with MustAdd and MustHave.
// Must operations panic on err caused by hash function, e.g., failed to convert hex to decimal
// (that shouldn't happen normally).
func Example_optimistic() {
	const maxEmails = 100
	const prob = 0.01
	bf, err := bloom.New(maxEmails, prob)
	if err != nil {
		log.Fatalf("Bloom filter is not created: %v", err)
	}

	email := []byte("alice@example.com")
	bf.MustAdd(email)
	if bf.MustHave(email) {
		fmt.Print("Alice's email possibly is in the set.")
	} else {
		fmt.Print("Alice's email definitely is not in the set.")
	}
	// Output:
	// Alice's email possibly is in the set.
}

// New returns the following errors if number of elements or probability are out of range.
func ExampleNew_error() {
	_, err := bloom.New(0, 0.01)
	if err == bloom.ErrZeroElements {
		fmt.Println(err)
	}

	_, err = bloom.New(1000000, 0)
	if err == bloom.ErrProbability {
		fmt.Println(err)
	}
	// Output:
	// number of elements must be positive
	// probability must be positive
}
