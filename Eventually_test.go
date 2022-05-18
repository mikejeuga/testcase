package testcase_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/random"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/internal"
)

func TestEventually(t *testing.T) {
	SpecEventually(t)
}

func BenchmarkEventually(b *testing.B) {
	SpecEventually(b)
}

func SpecEventually(tb testing.TB) {
	s := testcase.NewSpec(tb)

	var (
		strategyWillRetry = testcase.Var[bool]{ID: `retry strategy will retry`}
		strategyStub      = testcase.Let(s, func(t *testcase.T) *stubRetryStrategy {
			return &stubRetryStrategy{ShouldRetry: strategyWillRetry.Get(t)}
		})
		helper = testcase.Let(s, func(t *testcase.T) *testcase.Eventually {
			return &testcase.Eventually{
				RetryStrategy: strategyStub.Get(t),
			}
		})
	)

	s.Describe(`.Assert`, func(s *testcase.Spec) {
		var (
			stubTB  = testcase.Let(s, func(t *testcase.T) *testcase.StubTB { return &testcase.StubTB{} })
			blk     = testcase.Let(s, func(t *testcase.T) func(assert.It) { return func(it assert.It) {} })
			subject = func(t *testcase.T) {
				helper.Get(t).Assert(stubTB.Get(t), blk.Get(t))
			}
		)

		var (
			blkCounter     = testcase.LetValue(s, 0)
			blkCounterGet  = func(t *testcase.T) int { return blkCounter.Get(t) }
			blkDuration    = testcase.LetValue(s, time.Duration(0))
			blkDurationGet = func(t *testcase.T) time.Duration { return blkDuration.Get(t) }
			blkLet         = func(s *testcase.Spec, fn func(*testcase.T, testing.TB)) {
				blkCounterInc := func(t *testcase.T) { blkCounter.Set(t, blkCounter.Get(t)+1) }

				blk.Let(s, func(t *testcase.T) func(assert.It) {
					return func(tb assert.It) {
						blkCounterInc(t)
						time.Sleep(blkDurationGet(t))
						fn(t, tb)
					}
				})
			}
		)

		s.When(`the assertion fails`, func(s *testcase.Spec) {
			//s.Before(func(t *testcase.T) { t.Skip() }) // TODO

			blkLet(s, func(t *testcase.T, tb testing.TB) { tb.Fail() })

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := testcase.LetValue(s, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t)+1) })
						tb.Error(`foo`)
						tb.Errorf(`%s`, `baz`)
					})

					stubTB.Let(s, func(t *testcase.T) *testcase.StubTB {
						stub := &testcase.StubTB{}
						t.Cleanup(func() {
							t.Must.Contain(stub.Logs, `foo`)
							t.Must.Contain(stub.Logs, `baz`)
						})
						t.Cleanup(stub.Finish)
						return stub
					})

					s.Then(`list events replied to the passed testing.TB`, func(t *testcase.T) {
						subject(t)
					})

					s.Then(`cleanup is forwarded regardless the failed error`, func(t *testcase.T) {
						subject(t)

						t.Must.True(0 < cuCounter.Get(t))
					})
				})
			}

			s.And(`strategy don't allow further retries other than the first evaluation`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, false)

				s.Then(`it will execute the assertion at least once`, func(t *testcase.T) {
					subject(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`strategy will allow further retries even over the fist assertion block evaluation`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, true)

				s.Then(`it will run for as long as the wait timeout duration`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(strategyStub.Get(t).IsMaxReached())
				})

				s.Then(`it will execute the condition multiple times`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(1 < blkCounterGet(t))
				})

				s.Then(`it will fail the test`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.And(`it fails with FailNow`, func(s *testcase.Spec) {
					hasRun := testcase.LetValue(s, false)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() { hasRun.Set(t, true) })
						tb.FailNow()
					})

					s.Then(`it will fail the test`, func(t *testcase.T) {
						internal.Recover(func() { subject(t) })

						assert.Must(t).True(stubTB.Get(t).Failed())
					})

					s.Then(`it will ensure that Cleanup was executed`, func(t *testcase.T) {
						internal.RecoverExceptGoexit(func() { subject(t) })

						assert.Must(t).True(hasRun.Get(t))
					})
				})
			})
		})

		s.When(`the assertion returns with list happy`, func(s *testcase.Spec) {
			blkLet(s, func(t *testcase.T, tb testing.TB) {
				// nothing to do, TB then will not fail // tb.Success
			})

			andMultipleAssertionEventSentToTestingTB := func(s *testcase.Spec) {
				s.And(`and multiple assertion event sent to testing.TB`, func(s *testcase.Spec) {
					cuCounter := testcase.LetValue(s, 0)

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Log(`foo`)
						tb.Logf(`%s - %s`, `bar`, `baz`)
						tb.Cleanup(func() { cuCounter.Set(t, cuCounter.Get(t)+1) })
					})

					stubTB.Let(s, func(t *testcase.T) *testcase.StubTB {
						stub := &testcase.StubTB{}
						t.Cleanup(stub.Finish)
						t.Cleanup(func() {
							t.Must.Contain(stub.Logs, "foo")
							t.Must.Contain(stub.Logs, "bar - baz")
						})
						return stub
					})

					s.Then(`list events replied to the passed testing.TB`, func(t *testcase.T) {
						subject(t)
					})

					s.Then(`cleanup is forwarded`, func(t *testcase.T) {
						subject(t)
						stubTB.Get(t).Finish()
						t.Must.True(0 < cuCounter.Get(t))
					})
				})
			}

			s.And(`strategy will not retry the assertion block after the first execution`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, false)

				s.Then(`it will execute the condition at least once`, func(t *testcase.T) {
					subject(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(!stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)
			})

			s.And(`strategy allow multiple condition`, func(s *testcase.Spec) {
				strategyWillRetry.LetValue(s, true)

				s.Then(`it will not use up list the retry strategy loop iterations because the condition doesn't need it`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(!strategyStub.Get(t).IsMaxReached())
				})

				s.Then(`it will execute the condition only for the required required amount of times`, func(t *testcase.T) {
					subject(t)

					t.Must.Equal(1, blkCounterGet(t))
				})

				s.Then(`it will not mark the passed TB as failed`, func(t *testcase.T) {
					subject(t)

					assert.Must(t).True(!stubTB.Get(t).Failed())
				})

				andMultipleAssertionEventSentToTestingTB(s)

				s.Context(`but it will fail during the Cleanup`, func(s *testcase.Spec) {
					stubTB.Let(s, func(t *testcase.T) *testcase.StubTB {
						return &testcase.StubTB{}
					})

					blkLet(s, func(t *testcase.T, tb testing.TB) {
						tb.Cleanup(func() {
							tb.Logf(`I'm running and I'm pointing to %T`, tb)
							tb.FailNow()
						})
					})

					s.Then(`then cleanup is replied to the test subject`, func(t *testcase.T) {
						subject(t) // assertion in the TB mock
					})
				})

				s.And(`assertion fails a few times but then yields success`, func(s *testcase.Spec) {
					stubTB.Let(s, func(t *testcase.T) *testcase.StubTB {
						stub := &testcase.StubTB{}
						t.Cleanup(stub.Finish)
						t.Cleanup(func() {
							t.Must.False(stub.IsFailed)
						})
						return stub
					})

					var (
						cleanups       = testcase.Let(s, func(t *testcase.T) []string { return []string{} })
						cleanupsAppend = func(t *testcase.T, v ...string) {
							cleanups.Set(t, append(cleanups.Get(t), v...))
						}
					)
					blkLet(s, func(t *testcase.T, tb testing.TB) {
						t.Log(`ent`)
						tb.Cleanup(func() { cleanupsAppend(t, `foo`) })
						tb.Cleanup(func() { cleanupsAppend(t, `bar`) })
						tb.Cleanup(func() { cleanupsAppend(t, `baz`) })

						t.Log(`in`)
						// fail happens after the cleanups intentionally
						t.Log(`blkCounterGet`, blkCounterGet(t))
						if i := blkCounterGet(t); i < 3 {
							t.Log(`err`)
							tb.FailNow()
						}
						t.Log(`orderingOutput`)
					})

					s.Then(`failed runs cleanup after themselves`, func(t *testcase.T) {
						subject(t) // expectations in in the TB input as mock

						expected := []string{
							`baz`, `bar`, `foo`, // block runs first
							`baz`, `bar`, `foo`, // block runs for the second time
						}

						t.Must.Equal(expected, cleanups.Get(t))
					})
				})
			})
		})
	})

	s.Describe(`implements SpecOption`, func(s *testcase.Spec) {
		//subject := func() {}
	})
}

func TestRetry_Assert_failsOnceButThenPass(t *testing.T) {
	w := testcase.Eventually{
		RetryStrategy: testcase.Waiter{
			WaitDuration: 0,
			Timeout:      42 * time.Second,
		},
	}

	var (
		stub    = &testcase.StubTB{}
		counter int
		times   int
	)
	w.Assert(stub, func(it assert.It) {
		// run 42 times
		// 41 times it will fail but create cleanup
		// on the last it should pass
		//
		it.Cleanup(func() { counter++ })
		if 41 <= times {
			return
		}
		times++
		it.Fail()
	})

	t.Log("it is a design decision that the last cleanup is not executed during the assert looping")
	t.Log("the value might be still expected to be used.")
	assert.Must(t).Equal(41, counter)

	stub.Finish()
	assert.Must(t).Equal(42, counter)
}

func TestRetry_Assert_panic(t *testing.T) {
	w := testcase.Eventually{
		RetryStrategy: testcase.RetryStrategyFunc(func(condition func() bool) {
			for condition() {
			}
		}),
	}

	rnd := random.New(random.CryptoSeed{})
	expectedPanicValue := rnd.String()
	actualPanicValue := func() (r interface{}) {
		defer func() { r = recover() }()
		w.Assert(t, func(it assert.It) {
			panic(expectedPanicValue)
		})
		return nil
	}()

	assert.Must(t).Equal(expectedPanicValue, actualPanicValue)
}

type stubRetryStrategy struct {
	ShouldRetry bool
	counter     int
}

func (s *stubRetryStrategy) IsMaxReached() bool {
	return 42 <= s.counter
}

func (s *stubRetryStrategy) inc() bool {
	s.counter++

	return !s.IsMaxReached()
}

func (s *stubRetryStrategy) While(condition func() bool) {
	for condition() && s.inc() && s.ShouldRetry {
	}
}

func TestRetryCount_While(t *testing.T) {
	s := testcase.NewSpec(t)

	var (
		i        = testcase.Var[int]{ID: `max times`}
		strategy = testcase.Let(s, func(t *testcase.T) testcase.RetryStrategy {
			return testcase.RetryCount(i.Get(t))
		})
		condition = testcase.Var[bool]{ID: `condition`}
		subject   = func(t *testcase.T) int {
			var count int
			strategy.Get(t).While(func() bool {
				count++
				return condition.Get(t)
			})
			return count
		}
	)

	s.When(`max times is 0`, func(s *testcase.Spec) {
		i.LetValue(s, 0)

		s.And(`condition always yields true`, func(s *testcase.Spec) {
			condition.LetValue(s, true)

			s.Then(`it should run at least one times`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})

		s.And(`condition always yields false`, func(s *testcase.Spec) {
			condition.LetValue(s, false)

			s.Then(`it should stop on the first iteration`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})
	})

	s.When(`max times is greater than 0`, func(s *testcase.Spec) {
		i.Let(s, func(t *testcase.T) int {
			return random.New(random.CryptoSeed{}).IntBetween(1, 10)
		})

		s.And(`condition always yields true`, func(s *testcase.Spec) {
			condition.LetValue(s, true)

			s.Then(`it should run for the maximum retry count plus one for the initial run`, func(t *testcase.T) {
				t.Must.Equal(i.Get(t)+1, subject(t))
			})
		})

		s.And(`condition always yields false`, func(s *testcase.Spec) {
			condition.LetValue(s, false)

			s.Then(`it should stop on the first iteration`, func(t *testcase.T) {
				t.Must.Equal(1, subject(t))
			})
		})
	})
}