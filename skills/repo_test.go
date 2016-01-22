// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package skills

import (
	. "gopkg.in/check.v1"
)

type RepositorySuite struct {
	t         Type
	skill     *Skill
	emptyRepo *Repository
	// Repository pre-populated with s.t
	testRepo *Repository
}

var _ = Suite(&RepositorySuite{
	t: &TestType{
		TypeName: "type",
	},
	skill: &Skill{
		Snap:  "snap",
		Name:  "name",
		Type:  "type",
		Attrs: map[string]interface{}{"attr": "value"},
	},
})

func (s *RepositorySuite) SetUpTest(c *C) {
	s.emptyRepo = NewRepository()
	s.testRepo = NewRepository()
	err := s.testRepo.AddType(s.t)
	c.Assert(err, IsNil)
}

// Tests for Repository.AddType()

func (s *RepositorySuite) TestAddType(c *C) {
	// Adding a valid type works
	err := s.emptyRepo.AddType(s.t)
	c.Assert(err, IsNil)
	c.Assert(s.emptyRepo.Type(s.t.Name()), Equals, s.t)
	c.Assert(s.emptyRepo.AllTypes(), DeepEquals, []Type{s.t})
}

func (s *RepositorySuite) TestAddTypeClash(c *C) {
	t1 := &TestType{TypeName: "type"}
	t2 := &TestType{TypeName: "type"}
	err := s.emptyRepo.AddType(t1)
	c.Assert(err, IsNil)
	// Adding a type with the same name as another type is not allowed
	err = s.emptyRepo.AddType(t2)
	c.Assert(err, Equals, ErrDuplicate)
	c.Assert(s.emptyRepo.Type(t1.Name()), Equals, t1)
	c.Assert(s.emptyRepo.AllTypes(), DeepEquals, []Type{t1})
}

func (s *RepositorySuite) TestAddTypeInvalidName(c *C) {
	t := &TestType{TypeName: "bad-name-"}
	// Adding a type with invalid name is not allowed
	err := s.emptyRepo.AddType(t)
	c.Assert(err, ErrorMatches, `"bad-name-" is not a valid skill or slot name`)
	c.Assert(s.emptyRepo.Type(t.Name()), IsNil)
	c.Assert(s.emptyRepo.AllTypes(), HasLen, 0)
}

// Tests for Repository.Type()

func (s *RepositorySuite) TestType(c *C) {
	// Type returns nil when it cannot be found
	t := s.emptyRepo.Type(s.t.Name())
	c.Assert(t, IsNil)
	c.Assert(s.emptyRepo.Type(s.t.Name()), IsNil)
	err := s.emptyRepo.AddType(s.t)
	c.Assert(err, IsNil)
	// Type returns the found type
	t = s.emptyRepo.Type(s.t.Name())
	c.Assert(t, Equals, s.t)
}

func (s *RepositorySuite) TestTypeSearch(c *C) {
	err := s.emptyRepo.AddType(&TestType{TypeName: "a"})
	c.Assert(err, IsNil)
	err = s.emptyRepo.AddType(&TestType{TypeName: "b"})
	c.Assert(err, IsNil)
	err = s.emptyRepo.AddType(&TestType{TypeName: "c"})
	c.Assert(err, IsNil)
	// Type correctly finds types
	c.Assert(s.emptyRepo.Type("a"), Not(IsNil))
	c.Assert(s.emptyRepo.Type("b"), Not(IsNil))
	c.Assert(s.emptyRepo.Type("c"), Not(IsNil))
}

// Tests for Repository.AllTypes()

func (s *RepositorySuite) TestAllTypes(c *C) {
	tA := &TestType{TypeName: "a"}
	tB := &TestType{TypeName: "b"}
	tC := &TestType{TypeName: "c"}
	// Note added in non-sorted order
	err := s.emptyRepo.AddType(tA)
	c.Assert(err, IsNil)
	err = s.emptyRepo.AddType(tC)
	c.Assert(err, IsNil)
	err = s.emptyRepo.AddType(tB)
	c.Assert(err, IsNil)
	// All types are returned. Types are ordered by Name
	c.Assert(s.emptyRepo.AllTypes(), DeepEquals, []Type{tA, tB, tC})
}

// Tests for Repository.AddSkill()

func (s *RepositorySuite) TestAddSkill(c *C) {
	c.Assert(s.testRepo.AllSkills(""), HasLen, 0)
	err := s.testRepo.AddSkill(s.skill.Snap, s.skill.Name, s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	c.Assert(s.testRepo.AllSkills(""), HasLen, 1)
	c.Assert(s.testRepo.Skill(s.skill.Snap, s.skill.Name), DeepEquals, s.skill)
}

func (s *RepositorySuite) TestAddSkillClash(c *C) {
	err := s.testRepo.AddSkill(s.skill.Snap, s.skill.Name, s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill(s.skill.Snap, s.skill.Name, s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, Equals, ErrDuplicate)
	c.Assert(s.testRepo.AllSkills(""), HasLen, 1)
	c.Assert(s.testRepo.Skill(s.skill.Snap, s.skill.Name), DeepEquals, s.skill)
}

func (s *RepositorySuite) TestAddSkillInvalidName(c *C) {
	err := s.testRepo.AddSkill("snap", "bad-name-", "type", "label", nil)
	c.Assert(err, ErrorMatches, `"bad-name-" is not a valid skill or slot name`)
	c.Assert(s.testRepo.AllSkills(""), HasLen, 0)
}

// Tests for Repository.Skill()

func (s *RepositorySuite) TestSkill(c *C) {
	err := s.testRepo.AddSkill(s.skill.Snap, s.skill.Name, s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	c.Assert(s.emptyRepo.Skill(s.skill.Snap, s.skill.Name), IsNil)
	c.Assert(s.testRepo.Skill(s.skill.Snap, s.skill.Name), DeepEquals, s.skill)
}

func (s *RepositorySuite) TestSkillSearch(c *C) {
	err := s.testRepo.AddSkill("x", "a", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("x", "b", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("x", "c", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("y", "a", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("y", "b", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("y", "c", s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	// Skill() correctly finds skills
	c.Assert(s.testRepo.Skill("x", "a"), Not(IsNil))
	c.Assert(s.testRepo.Skill("x", "b"), Not(IsNil))
	c.Assert(s.testRepo.Skill("x", "c"), Not(IsNil))
	c.Assert(s.testRepo.Skill("y", "a"), Not(IsNil))
	c.Assert(s.testRepo.Skill("y", "b"), Not(IsNil))
	c.Assert(s.testRepo.Skill("y", "c"), Not(IsNil))
}

// Tests for Repository.RemoveSkill()

func (s *RepositorySuite) TestRemoveSkillGood(c *C) {
	err := s.testRepo.AddSkill(s.skill.Snap, s.skill.Name, s.skill.Type, s.skill.Label, s.skill.Attrs)
	c.Assert(err, IsNil)
	err = s.testRepo.RemoveSkill(s.skill.Snap, s.skill.Name)
	c.Assert(err, IsNil)
	c.Assert(s.testRepo.AllSkills(""), HasLen, 0)
}

func (s *RepositorySuite) TestRemoveSkillNoSuchSkill(c *C) {
	err := s.emptyRepo.RemoveSkill(s.skill.Snap, s.skill.Name)
	c.Assert(err, Equals, ErrSkillNotFound)
}

// Tests for Repository.AllSkills()

func (s *RepositorySuite) TestAllSkillsWithoutTypeName(c *C) {
	// Note added in non-sorted order
	err := s.testRepo.AddSkill("snap-b", "name-a", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-b", "name-c", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-b", "name-b", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-a", "name-a", "type", "label", nil)
	c.Assert(err, IsNil)
	// The result is sorted by snap and name
	c.Assert(s.testRepo.AllSkills(""), DeepEquals, []*Skill{
		&Skill{
			Snap:  "snap-a",
			Name:  "name-a",
			Type:  "type",
			Label: "label",
		},
		&Skill{
			Snap:  "snap-b",
			Name:  "name-a",
			Type:  "type",
			Label: "label",
		},
		&Skill{
			Snap:  "snap-b",
			Name:  "name-b",
			Type:  "type",
			Label: "label",
		},
		&Skill{
			Snap:  "snap-b",
			Name:  "name-c",
			Type:  "type",
			Label: "label",
		},
	})
}

func (s *RepositorySuite) TestAllSkillsWithTypeName(c *C) {
	// Add another type so that we can look for it
	err := s.testRepo.AddType(&TestType{TypeName: "other-type"})
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap", "name-a", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap", "name-b", "other-type", "label", nil)
	c.Assert(err, IsNil)
	// The result is sorted by snap and name
	c.Assert(s.testRepo.AllSkills("other-type"), DeepEquals, []*Skill{
		&Skill{
			Snap:  "snap",
			Name:  "name-b",
			Type:  "other-type",
			Label: "label",
		},
	})
}

// Tests for Repository.Skills()

func (s *RepositorySuite) TestSkills(c *C) {
	// Note added in non-sorted order
	err := s.testRepo.AddSkill("snap-b", "name-a", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-b", "name-c", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-b", "name-b", "type", "label", nil)
	c.Assert(err, IsNil)
	err = s.testRepo.AddSkill("snap-a", "name-a", "type", "label", nil)
	c.Assert(err, IsNil)
	// The result is sorted by snap and name
	c.Assert(s.testRepo.Skills("snap-b"), DeepEquals, []*Skill{
		&Skill{
			Snap:  "snap-b",
			Name:  "name-a",
			Type:  "type",
			Label: "label",
		},
		&Skill{
			Snap:  "snap-b",
			Name:  "name-b",
			Type:  "type",
			Label: "label",
		},
		&Skill{
			Snap:  "snap-b",
			Name:  "name-c",
			Type:  "type",
			Label: "label",
		},
	})
	// The result is empty if the snap is not known
	c.Assert(s.testRepo.Skills("snap-x"), HasLen, 0)
}
